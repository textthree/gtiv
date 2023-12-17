package dto

type VideoHomeItem struct {
	Id         int `v:"required"`
	UserId     int
	Title      string `v:"required"`
	Uri        string `v:"required"`
	Cover      string `v:"required"`
	Width      int    `v:"required"`
	Height     int    `v:"required"`
	Nickname   string
	Avatar     string
	IsFollow   bool
	SupportNum int
	CollectNum int
	ShareNum   int
}

type Response struct {
	Code    int
	Message string
}

type VideoHomeRes struct {
	BaseRes
	Endpoint string          `v:"required"`
	List     []VideoHomeItem `v:"required"`
}

type UserHomeVideoItem struct {
	VideoId    int
	Title      string
	Cover      string
	SupportNum int
}
type UserHomeVideoListRes struct {
	BaseRes
	Endpoint string
	List     []UserHomeVideoItem
}

type VideoInfo struct {
	MasterUid    int
	MasterAvatar string
	IsFollow     bool
	SupportNum   int
	CollectNum   int
	ShareNum     int
	VideoCover   string
	VideoWidth   int
	VideoHeight  int
}
