package dto

type CountryReq struct {
}
type CountryItemRes struct {
	TelCode int    // 国际区号
	Country string // 国家
}
type CountryRes struct {
	BaseRes
	List []CountryItemRes
}

type VersionReq struct {
	// 上传类型：1.安卓 2.ios
	Type int
}

type VersionRes struct {
	BaseRes
	Version      string
	Url          string
	ForceUpgrade bool
}
