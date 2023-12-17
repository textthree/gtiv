package video

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"github.com/nfnt/resize"
	"github.com/spf13/cast"
	"github.com/text3cn/goodle/kit/filekit"
	"github.com/text3cn/goodle/kit/gokit"
	"github.com/text3cn/goodle/providers/goodlog"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"gtiv/app/videobiz/entity"
	"gtiv/app/videobiz/internal/boot"
	"gtiv/kit/ffmpeg"
	"gtiv/kit/obs"
	"image"
	"image/jpeg"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"
)

var endpoint string
var azure *obs.AzureClient
var connStr string
var containName = "videos"

func UploadVideo(ctx *httpserver.Context) {
	var limit int64 = 10 // 限制上传大小 10 M
	tempFile, handle, params, err := ctx.Req.FormFile("file")
	if err != nil {
		msg := "超过文件大小限制，最大允许上传 " + cast.ToString(limit) + " MB"
		goodlog.Error(msg)
		return
	}
	qiniu := obs.QiniuClient{
		AK:       boot.VideobizCfg.GetString("qiniuObs.accessKey"),
		SK:       boot.VideobizCfg.GetString("qiniuObs.secretKey"),
		BuckName: boot.VideobizCfg.GetString("qiniuObs.videoBucketName"),
	}
	// 获取文件 hash
	fileData, err := io.ReadAll(tempFile)
	hash := sha1.New()
	hash.Write(fileData)
	hashValue := hex.EncodeToString(hash.Sum(nil))
	suffix := filekit.GetSuffix(handle.Filename)
	// 文件保存到本地
	tempDir := "./upload_temp/"
	workDir := path.Join(tempDir, hashValue)
	filekit.MkDir(workDir, 0755)
	videoFileLocalPath := path.Join(workDir, hashValue) + "." + suffix
	f, err := os.OpenFile(videoFileLocalPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	reader := bytes.NewReader(fileData)
	io.Copy(f, reader)
	// 截取封面上传
	obsdir := cast.ToString(time.Now().Year()) + "/" + hashValue + "/"
	coverObskey := obsdir + "cover.jpg"
	go gokit.SafeGo(func() {
		cover := path.Join(workDir, hashValue) + ".jpg"
		ffmpeg.GetVideoCover(videoFileLocalPath, cover)
		// 压缩封面
		compressCover(cover)
		qiniu.UploadFromFile(cover, coverObskey)
		//coverConent, _ := os.ReadFile(cover)
		//err = client.AzureUploadBuffer(coverObskey, coverConent, "image/jpg")
		if err != nil {
			return
		}
	})
	// 提取视频元数据
	metaData := ffmpeg.GetMetaData(videoFileLocalPath)
	// 切片、上传
	tsWorkDir := path.Join(workDir, "ts")
	ffmpeg.TsFileCutH264(videoFileLocalPath, tsWorkDir, "3", "soft")
	uploadTs(&qiniu, tsWorkDir, obsdir)
	// 删除工作目录
	os.RemoveAll(tempDir + hashValue)
	// 保存到数据库
	videoObskey := obsdir + "index.m3u8"
	saveToDb(ctx, params, coverObskey, videoObskey, metaData)
	ctx.Resp.Json(map[string]string{
		"video": endpoint + videoObskey,
		"cover": endpoint + coverObskey,
	})
	return
}

// 重新上传某个视频，根据原来的 hashValue 重传
func ReUploadVideo(ctx *httpserver.Context) {
	var limit int64 = 10 // 限制上传大小 10 M
	tempFile, handle, params, err := ctx.Req.FormFile("file")
	if err != nil {
		msg := "超过文件大小限制，最大允许上传 " + cast.ToString(limit) + " MB"
		goodlog.Error(msg)
		return
	}
	qiniu := obs.QiniuClient{
		AK:       boot.VideobizCfg.GetString("qiniuObs.accessKey"),
		SK:       boot.VideobizCfg.GetString("qiniuObs.secretKey"),
		BuckName: boot.VideobizCfg.GetString("qiniuObs.videoBucketName"),
	}
	hashValue := params.Get("hashValue")
	obsdir := params.Get("obsDir")
	suffix := filekit.GetSuffix(handle.Filename)
	fileData, err := io.ReadAll(tempFile)
	// 文件保存到本地
	tempDir := "./temp/"
	workDir := path.Join(tempDir, hashValue)
	filekit.MkDir(workDir, 0755)
	saveFileName := path.Join(workDir, hashValue) + "." + suffix
	f, err := os.OpenFile(saveFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	reader := bytes.NewReader(fileData)
	io.Copy(f, reader)
	// 切片、上传
	tsWorkDir := path.Join(workDir, "ts")
	ffmpeg.TsFileCutH264(saveFileName, tsWorkDir, "3", "soft")
	uploadTs(&qiniu, tsWorkDir, obsdir)
	// 删除工作目录
	os.RemoveAll(tempDir + hashValue)
	ctx.Resp.Json("上传完成：" + endpoint + obsdir + "index.m3u8")
	return
}

// 塞进管道多协程同时上传
func uploadTs(qiniu *obs.QiniuClient, tsWorkDir, obsdir string) {
	files, _ := filekit.Scandir(tsWorkDir)
	tsQuantiy := len(files)
	if tsQuantiy < 1 {
		goodlog.Warn("没有切片")
		return
	}
	tsChan := make(chan string, tsQuantiy)
	exitChan := make(chan bool, tsQuantiy)
	for _, v := range files {
		tsChan <- v
	}
	close(tsChan)
	uploadTsFile := func() {
		localTsFilePath, ok := <-tsChan
		if !ok {
			exitChan <- true
			return
		}
		//tsContent, err := os.ReadFile(file)
		//if err != nil {
		//	goodlog.Pink("切片读取异常")
		//	return
		//}
		_, fileName := filepath.Split(localTsFilePath)
		tsObjectName := obsdir + fileName
		goodlog.Pink(tsObjectName)
		var err error
		if fileName == "index.m3u8" {
			//err = client.AzureUploadBuffer(tsObjectName, tsContent, "application/x-mpegurl")
			err = qiniu.UploadFromFile(localTsFilePath, tsObjectName)
		} else {
			//err = client.AzureUploadBuffer(tsObjectName, tsContent)
			err = qiniu.UploadFromFile(localTsFilePath, tsObjectName)
		}
		if err != nil {
			goodlog.Error(err)
		}
		if err != nil {
			goodlog.Pink("切片上传异常")
		}
		exitChan <- true
	}
	// 切片后有多少个文件就开多少个协程
	for i := 1; i <= tsQuantiy; i++ {
		go gokit.SafeGo(func() {
			uploadTsFile()
		})
	}
	// 上传完毕
	finishGroutineCount := 0
	finishWait := false
	for {
		if finishWait == true {
			break
		}
		select {
		// 10 分钟后强制退出for循环，防止永久阻塞
		case <-time.After(10 * time.Minute):
			goodlog.Pink("10 分钟超时")
			finishWait = true
		case _, ok := <-exitChan:
			if ok {
				finishGroutineCount++
			}
			if finishGroutineCount >= tsQuantiy {
				close(exitChan)
				finishWait = true
			}
		}
	}
}

func saveToDb(ctx *httpserver.Context, params url.Values, coverUri, videoUri string, metaData ffmpeg.MetaData) {
	now := time.Now()
	userId := params.Get("userId")
	title := params.Get("title")
	video := entity.Video{
		UserId:      cast.ToUint(userId),
		Title:       title,
		Cover:       coverUri,
		Uri:         videoUri,
		Size:        int(metaData.Size / 1024),
		Width:       uint16(metaData.Width),
		Height:      uint16(metaData.Height),
		Duration:    int32(metaData.Duration),
		BitRate:     uint32(metaData.BitRate / 1024),
		CreatedTime: now,
		UpdatedTime: now,
	}
	err := orm.GetDB().Create(&video).Error
	if err != nil {
		goodlog.Error(err)
	} else {
		goodlog.Info("入库完成")
	}
}

func compressCover(cover string) {
	// 打开图像文件
	file, err := os.Open(cover)
	if err != nil {
		goodlog.Error(err)
	}

	// 解码图像文件
	img, _, err := image.Decode(file)
	if err != nil {
		goodlog.Error(err)
	}

	// 指定目标宽度（等比例缩放）
	targetWidth := 800 // 目标宽度
	ratio := float64(targetWidth) / float64(img.Bounds().Dx())
	targetHeight := uint(float64(img.Bounds().Dy()) * ratio)
	resizedImg := resize.Resize(uint(targetWidth), targetHeight, img, resize.Lanczos3)

	// 创建输出文件
	output, err := os.Create(cover)
	if err != nil {
		goodlog.Error(err)
	}

	// 将压缩后的图像保存到输出文件
	err = jpeg.Encode(output, resizedImg, nil)
	if err != nil {
		goodlog.Error(err)
	}

	goodlog.Trace("图像已压缩", cover)
}
