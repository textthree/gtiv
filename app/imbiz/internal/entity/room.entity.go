package entity

import "time"

type Room struct {
	Id          uint `gorm:"autoIncrement"`
	RoomName    string
	Avatar      string
	MemberLimit int       `gorm:"comment:限制人数;not null; default:500"`
	Notice      string    `gorm:"type:text;comment:群公告"`
	MemberNum   int       `gorm:"comment:群人数;default:0;not null"`
	CreateTime  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdateTime  time.Time `gorm:"default:null"`
}
