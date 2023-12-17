package entity

import (
	"time"
)

type Follow struct {
	Id         uint `gorm:"autoIncrement"`
	Master     int  `gorm:"uniqueIndex:idx_master_fans;index;default:0;not null;comment:被关注者用户 id"`
	Fans       int  `gorm:"uniqueIndex:idx_master_fans;default:0;not null;comment:关注者用户 id"`
	CreateTime time.Time
}
