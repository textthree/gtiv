package obs

//
//import (
//	"bytes"
//	"fmt"
//)
//
//type huaweiObs struct {
//	bucketName string
//	obsClient  *obs.ObsClient
//}
//
//// 初始化
//func (this *huaweiObs) init() {
//	endpoint := config.Get.HuaweiCloud.ObsEndpoint
//	this.bucketName = config.Get.HuaweiCloud.BucketName
//	ak := config.Get.HuaweiCloud.ObsAccessKey
//	sk := config.Get.HuaweiCloud.ObsSecretKey
//	// 创建 ObsClient 结构体
//	obsClient, err := obs.New(ak, sk, endpoint)
//	if err != nil {
//		logger.Error("[init bosClient] 华为 SDK 初始化 obsClient 失败 " + err.Error())
//	}
//	this.obsClient = obsClient
//}
//
//// 上传
//// fileConent 文件内容
//// objectName 文件名称,包含文件存储路径，如：2021/05/xxx.jpg
//// args[0] content-type
//func (this huaweiObs) obsUploadOne(fileConent []byte, objectName string, contentType string) (error errortype.BizError) {
//	this.init()
//	input := &obs.PutObjectInput{}
//	input.Bucket = this.bucketName
//	input.Key = objectName
//	input.Body = bytes.NewReader(fileConent) // 字节数组转 io.Reader（文件）
//	if contentType != "" {
//		input.ContentType = contentType
//	}
//	_, err := this.obsClient.PutObject(input) // output, err := ......
//	if err != nil {
//		error.Content = "华为云上传失败"
//		error.Errors = err
//		error.Code = errorcodes.HUAWEI_BOS_UPLOAD_FAILED
//	}
//	//fmt.Printf("RequestId:%s\n", output.RequestId)
//	//fmt.Printf("ETag:%s\n", output.ETag)
//	// fmt.Printf("File:%s\n", objectName)
//	// 关闭obsClient
//	this.obsClient.Close()
//	return
//}
//
//// 下载（获取内容）
//func (this huaweiObs) ObsGetOne(objectName string) string {
//	this.init()
//	input := &obs.GetObjectInput{}
//	input.Bucket = this.bucketName
//	input.Key = objectName
//	output, err := this.obsClient.GetObject(input)
//	var ret string
//	if err == nil {
//		defer output.Body.Close()
//		p := make([]byte, 1024)
//		var readErr error
//		var readCount int
//		// 读取对象内容
//		for {
//			readCount, readErr = output.Body.Read(p)
//			if readCount > 0 {
//				//fmt.Printf("%s", p[:readCount])
//				ret += string(p[:readCount])
//			}
//			if readErr != nil {
//				break
//			}
//		}
//	} else if obsError, ok := err.(obs.ObsError); ok {
//		fmt.Printf("Code:%s\n", obsError.Code)
//		fmt.Printf("Content:%s\n", obsError.Content)
//	}
//
//	return ret
//}
//
//// 删除
//func (this huaweiObs) ObsDelete(objectName string) {
//	this.init()
//	input := &obs.DeleteObjectInput{}
//	input.Bucket = this.bucketName
//	input.Key = objectName
//	_, err := this.obsClient.DeleteObject(input)
//	if err != nil {
//		if obsError, ok := err.(obs.ObsError); ok {
//			fmt.Println(obsError.Code)
//			fmt.Println(obsError.Content)
//		} else {
//			fmt.Println(err)
//		}
//	}
//}
