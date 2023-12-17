package api

import (
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/text3cn/goodle/providers/httpserver"
	"gtiv/app/imbiz/boot"
	"gtiv/app/imbiz/internal/constants"
	"gtiv/app/imbiz/internal/dto"
)

// @Summary 获取七牛 Token
// @Tags app
// @Router /obstore/client_token [get]
// @Param request body dto.ClientTokenReq true " "
// @Success 200 {object} dto.ClientTokenRes
func ClientToken(ctx *httpserver.Context) {
	// IM 聊天资源。7 天有效，非聊天资源，永久有效。
	var bucket, remoteBaseUrl string
	req := dto.ClientTokenReq{}
	ctx.Req.JsonScan(&req)
	switch req.Type {
	case constants.OBS_BUCKET_CHAT:
		bucket = boot.ImbizCfg.GetString("qiniuObs.chatBucketName")
		remoteBaseUrl = boot.ImbizCfg.GetString("qiniuObs.chatBucketUrl")
	case constants.OBS_BUCKET_IM:
		bucket = boot.ImbizCfg.GetString("qiniuObs.imBucketName")
		remoteBaseUrl = boot.ImbizCfg.GetString("qiniuObs.imBucketUrl")
	case constants.OBS_BUCKET_VIDEO:
		bucket = boot.ImbizCfg.GetString("qiniuObs.videoBucketName")
		remoteBaseUrl = boot.ImbizCfg.GetString("qiniuObs.videoBucketUrl")
	}
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	putPolicy.Expires = 7200 // token 2 小时有效
	// 视频处理
	/*if req.Type == constants.OBS_BUCKET_VIDEO {
		// https://developer.qiniu.com/dora/1313/video-frame-thumbnails-vframe
		saveJpgTo := base64.URLEncoding.EncodeToString([]byte(bucket + ":" + req.ObjectKey + ".jpg"))
		putPolicy.PersistentOps = "vframe/jpg/offset/3/w/300|saveas/" + saveJpgTo
	}*/

	ak := boot.ImbizCfg.GetString("qiniuObs.accessKey")
	sk := boot.ImbizCfg.GetString("qiniuObs.secretKey")
	mac := qbox.NewMac(ak, sk)
	upToken := putPolicy.UploadToken(mac)
	data := dto.ClientTokenRes{
		ImObsToken:    upToken,
		RemoteBaseUrl: remoteBaseUrl,
	}
	ctx.Resp.Json(data)
}

// 七牛视频截图
// http://rfgqrlavl.hn-bkt.clouddn.com/b1bf20d2-eeea-775d-8e52-0e792f1a79bd
/*func (c *objectStoreCtl) VideoThumb(ctx context.Context, req *v1.VideoThumbReq) (
	res *v1.VideoThumbRes, err error) {

	bucket := config.BizCfg.ObjectStore.AccessKey
	var url = "https://api.qiniu.com/pfop/" + "bucket=" + config.BizCfg.ObjectStore.Im.BucketName +
		"&key=" + req.ObjectKey +
		"&fops=vframe%2fjpg%2foffset%2f7%2fw%2f480%2fh%2f360"
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	putPolicy.Expires = 7200 // token 2 小时有效
	mac := qbox.NewMac(bucket, config.BizCfg.ObjectStore.SecretKey)
	token := putPolicy.UploadToken(mac)
	result, err := httpkit.Post(url, map[string]string{}, map[string]string{
		"Content-MsgType":  "application/x-www-form-urlencoded",
		"Authorization": token,
	})
	if err != nil {
		logger.Cblue(err)
	}
	echo.P(result)
	return
}
*/
