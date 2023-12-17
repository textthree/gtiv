package ffmpeg

import (
	"fmt"
	"github.com/text3cn/goodle/kit/filekit"
	"github.com/text3cn/goodle/kit/mathkit"
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/kit/syskit"
	"github.com/text3cn/goodle/kit/typekit"
	"github.com/text3cn/goodle/providers/goodlog"
	_ "image/jpeg" // 导入用于处理 JPEG 格式的图像
	"os"
	"os/exec"
	"path"
)

// 视频元数据
type MetaData struct {
	Duration int   // 持续时长，单位秒
	Size     int64 // 文件大小，单位字节
	BitRate  int64 // 码率
	Width    int   // 宽
	Height   int   // 高
}

var (
	ffmpegPath  string
	ffprobePath string
)

func init() {
	dir, _ := os.Getwd()
	ffmpegPath = dir + "/ffmpeg"
	ffprobePath = dir + "/ffprobe"
}

// 截取视频封面
func GetVideoCover(videoPath string, imageSavePath string) bool {
	cmdArgs := []string{
		"-i", videoPath,
		"-ss", "00:00:0.000", // -ss：起始时间，如 -ss 01:30:14，从 01:30:14 开始
		"-vframes", "1", // 指定截取的帧数，如-vframes 10，指定截取 10 张
		imageSavePath,
		"-y", // -y 覆盖保存
	}
	cmd := exec.Command(ffmpegPath, cmdArgs...)
	err := cmd.Run()
	if err != nil {
		goodlog.Error("截取视频封面错误：", err.Error(), cmd.Args)
		out, _ := cmd.CombinedOutput()
		goodlog.Pink(string(out))
		return false
	}
	return true
}

// 截取视频封面（docker运行）
//func GetVideoCover(videoPath string, imageSavePath string) bool {
//	cmdArgs := []string{
//		"run", "-v",
//		T3Config.Get.Media.DockerVolumes,
//		"--cpuset-cpus=" + T3Config.Get.Media.DockerCpus, "-c", "512", "-m",
//		T3Config.Get.Media.DockerMemory,
//		T3Config.Get.Media.DockerImage,
//		"ffmpeg", "-i",
//		videoPath, "-ss", "00:00:0.000", "-vframes", "1", imageSavePath, "-y"}
//	cmd := exec.Command("docker", cmdArgs...)
//	err := cmd.Run()
//	if err != nil {
//		logger.Errorf("截取视频封面错误：", err.Error())
//		fmt.Println(cmd.Args)
//		out, _ := cmd.CombinedOutput()
//		fmt.Println(string(out))
//		return false
//	}
//	return true
//}

// 提取元数据
func GetMetaData(filePath string) (metaData MetaData) {
	cmdArgs := []string{"-of", "json", "-show_format", filePath}
	_, output := syskit.ExecSysCmd(ffprobePath, cmdArgs)
	res, _ := typekit.JsonDecodeMap(output)
	res = res["format"].(map[string]interface{})
	metaData.Size = strkit.ParseInt64(res["size"].(string))
	duration := strkit.StringToFloat64(res["duration"].(string))
	metaData.Duration = int(mathkit.Round(duration))
	metaData.BitRate = strkit.ParseInt64(res["bit_rate"].(string))
	cmdArgs = []string{"-of", "json", "-show_streams", filePath}
	_, output = syskit.ExecSysCmd(ffprobePath, cmdArgs)
	res, _ = typekit.JsonDecodeMap(output)
	res2 := res["streams"].([]interface{})
	mp2 := res2[0].(map[string]interface{})
	if _, ok := mp2["height"].(float64); ok {
		metaData.Height = int(mp2["height"].(float64))
		metaData.Width = int(mp2["width"].(float64))
	} else {
		mp2 = res2[1].(map[string]interface{})
		if _, ok := mp2["height"].(float64); ok {
			metaData.Height = int(mp2["height"].(float64))
			metaData.Width = int(mp2["width"].(float64))
		}
	}
	return
}

// 视频切片成ts文件, h264编码
// filePath 被切的视频
// outputPath 切出来放哪
// hlsTime  分隔成几秒钟一个文件，ffmpeg 默认是 4 秒
// mode 编码格式，hard N卡硬编码，soft 软编码
func TsFileCutH264(filePath, outputPath string, hlsTime string, mode string) {
	filekit.MkDir(outputPath, 0755)
	m3u8 := path.Join(outputPath, "index.m3u8")
	var cmdArgs []string
	if mode == "soft" {
		// 不加密
		cmdArgs = []string{
			"-i", filePath, // -i 即：input filePath
			"-c:v", "libx264", // 编码格式
			"-c:a", "aac",
			"-strict", "2",
			"-f", "hls",
			"-hls_list_size", "0",
			"-hls_time", hlsTime, m3u8, "-y"}
		/* 加密：$videokey = ROOT . "/application/config/videokey.info";
		   $commandForFull = "ffmpeg -i $waterVideo -c copy -bsf:v h264_mp4toannexb -hls_time 5 -hls_list_size 0 -hls_key_info_file $videokey $FullM3u8";
		*/
	} else if mode == "hard" {
		cmdArgs = []string{
			"-i", filePath,
			"-c:v", "h264_nvenc",
			"-c:a", "aac",
			"-strict", "2",
			"-f", "hls",
			"-hls_list_size", "0",
			"-preset:v", "fast",
			"-hls_time", hlsTime, m3u8, "-y"}
	}
	cmd := exec.Command(ffmpegPath, cmdArgs...)
	err := cmd.Run()
	if err != nil {
		goodlog.Error(".TS切片失败：", err.Error())
		fmt.Println(cmd.Args)
		out, _ := cmd.CombinedOutput()
		fmt.Println(string(out))
	}
}
