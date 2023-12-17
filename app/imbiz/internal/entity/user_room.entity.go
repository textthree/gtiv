package entity

import "time"

type UserRoom struct {
	Id         uint      `gorm:"autoIncrement"`
	UserId     int       `gorm:"uniqueIndex:idx_uidrid;index;default:0;not null"`
	RoomId     int       `gorm:"uniqueIndex:idx_uidrid;default:0;not null"`
	Role       int8      `gorm:"default:10;comment:用户角色 10.普通用户 20.系统管理员，啥权限都有 100.开发人员"`
	CreateTime time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdateTime time.Time `gorm:"default:null"`
}
