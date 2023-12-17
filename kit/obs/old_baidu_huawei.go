package obs

//
//// 上传一个文件到对象存储
//// param archiveType   附件类型 - 不同类型存储到不同路径
//// param fileConent    文件内容
//// param args[0]	   content-type
//// @param fileName     文件名，包含后缀
//// @return objectName  文件在对象存储中的完整地址
//func UploadOne(archiveType string, fileConent []byte, fileName string, args ...string) (errortype.BizError, string, string) {
//	objectPath, domain, errorMessage := GetSavePath(archiveType)
//	objectName := objectPath + fileName
//	if errorMessage != "" {
//		error := errortype.BizError{
//			Code:    errorcodes.FAILED,
//			Content: errorMessage,
//		}
//		return error, domain, objectName
//	}
//	contentType := ""
//	if len(args) > 0 {
//		contentType = args[0]
//	}
//	obs := huaweiObs{}
//	error := obs.obsUploadOne(fileConent, objectName, contentType)
//	/*bos := baiduBos{
//		bucketName : configs.Get.BaiduCloud.BucketName,
//	}
//	error := bos.uploadOne(fileConent, objectName)*/
//	return error, domain, objectName
//}
//
//// 根据附件类型分配存储路径
//// archiveType  附件类型
//// domain       对象访问域名，不同上传类型附件可能会存到不同桶
//func GetSavePath(archiveType string) (path string, domain string, errorMessage string) {
//	huawei := functions.ObsAgentDomain()
//	switch archiveType {
//	case consts.SPIDER_UPLOAD_ARTICLE_IMAGE:
//		path = "spd/image/"
//		domain = huawei
//		break
//	case consts.SPIDER_UPLOAD_ARTICLE_VIDEO:
//		path = "spd/video/"
//		domain = huawei
//		break
//	case consts.SPIDER_UPLOAD_ARTICLE_CONTENT:
//		path = "spd/content/"
//		domain = huawei
//	case consts.UGC_UPLOAD_IMAGE:
//		y, m, _ := funcs.DateTodayInt()
//		path = "ugc/images/" + funcs.Tostring(y) + funcs.Tostring(m) + "/"
//		domain = huawei
//	default:
//		errorMessage = "附件类型不正确"
//	}
//	return
//}
//
//// 删除对象
//func DeleteOne(objectName string) {
//	obs := huaweiObs{}
//	obs.ObsDelete(objectName)
//}
//
//// 获取对象
//// param objectName 存储路径
//func GetOne(objectName string) string {
//	obs := huaweiObs{}
//	return obs.ObsGetOne(objectName)
//}
