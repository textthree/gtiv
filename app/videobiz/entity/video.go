package entity

import (
	"time"
)

type Video struct {
	Id          uint
	UserId      uint
	Title       string
	Cover       string
	Uri         string
	Width       uint16
	Height      uint16
	BitRate     uint32 // 单位 kbps（千比特），即 bit 除以 1024
	Duration    int32
	Size        int // 单位 KB
	SupportNum  int
	CollectNum  int
	ShareNum    int
	CreatedTime time.Time
	UpdatedTime time.Time
}

func (Video) TableName() string {
	return "videos"
}
