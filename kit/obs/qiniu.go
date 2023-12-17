package obs

import (
	"context"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/text3cn/goodle/providers/goodlog"
)

// https://developer.qiniu.com/kodo/1238/go#server-upload

type QiniuClient struct {
	AK       string
	SK       string
	BuckName string
}

func (QiniuClient) getCfg() *storage.Config {
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Region = &storage.ZoneHuanan
	// 是否使用https域名
	cfg.UseHTTPS = true
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	return &cfg
}

func (self *QiniuClient) UploadFromFile(localFile, objectKey string) error {
	putPolicy := storage.PutPolicy{
		Scope: self.BuckName,
	}
	mac := qbox.NewMac(self.AK, self.SK)
	upToken := putPolicy.UploadToken(mac)
	// 构建表单上传的对象
	formUploader := storage.NewFormUploader(self.getCfg())
	// 可选配置
	putExtra := storage.PutExtra{}
	ret := storage.PutRet{}
	err := formUploader.PutFile(context.Background(), &ret, upToken, objectKey, localFile, &putExtra)
	if err != nil {
		goodlog.Error(err)
		return err
	}
	goodlog.Info(ret.Key)
	return nil
}
