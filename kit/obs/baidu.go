package obs

//
//type baiduBos struct {
//	bucketName string
//}
//
//// 初始化百度 SDK 得到句柄
//func (this *baiduBos) init() *bos.Client {
//	endpoint := config.Get.BaiduCloud.BosEndpoint
//	clientConfig := bos.BosClientConfiguration{
//		Ak:               config.Get.BaiduCloud.BosAccessKey,
//		Sk:               config.Get.BaiduCloud.BosSecretKey,
//		Endpoint:         endpoint,
//		RedirectDisabled: false,
//	}
//	bosClient, err := bos.NewClientWithConfig(&clientConfig)
//	if err != nil {
//		logger.Error("[init bosClient] 百度 SDK 初始化 BosClient 失败 " + err.Error())
//	}
//	return bosClient
//}
//
//// 上传
//// fileConent 文件内容
//// objectName 文件名称,包含文件存储路径，如：2021/05/xxx.jpg
//func (this baiduBos) uploadOne(fileConent []byte, objectName string) (error errortype.BizError) {
//	_, err := this.init().PutObjectFromBytes(this.bucketName, objectName, fileConent, nil)
//	if err != nil {
//		error.Content = "百度云上传失败"
//		error.Errors = err
//		error.Code = errorcodes.BAIDU_BOS_UPLOAD_FAILED
//	}
//	return
//}
