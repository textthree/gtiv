package entity

import (
	"time"
)

// 用户点赞视频记录
type VideoSupport struct {
	Id            uint
	UserId        uint `gorm:"index"`
	VideoId       string
	VideoMasterId int
	CreateTime    time.Time
}
