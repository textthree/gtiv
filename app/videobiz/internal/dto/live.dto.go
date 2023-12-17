package dto

type LiveListItem struct {
	RoomId   string `v:"required"`
	UserId   int    `v:"required"`
	Title    string `v:"required"`
	Cover    string `v:"required"`
	State    byte   `v:"required"` // 1.直播中 0.未上播
	Avatar   string
	Nickname string
	IsFollow bool // 是否已关注该用户
}
type LiveListRes struct {
	BaseRes
	Rtmp string
	Hls  string
	Flv  string
	List []LiveListItem `v:"required"`
}
