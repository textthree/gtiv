package entity

import (
	"time"
)

// 用户收藏视频记录
type VideoCollect struct {
	Id         uint
	UserId     uint `gorm:"index"`
	VideoId    string
	CreateTime time.Time
}
