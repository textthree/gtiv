package dto

type ClientTokenReq struct {
	// 上传类型：1.聊天资源（7 天过期） 2.非聊天资源（不过期）
	Type int
	// 对象key
	ObjectKey string
}

type ClientTokenRes struct {
	BaseRes
	ImObsToken string
	// 桶的资源外链域名
	RemoteBaseUrl string
}
