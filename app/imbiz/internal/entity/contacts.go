package entity

import "time"

type Contacts struct {
	Id             uint      `gorm:"autoIncrement"`
	UserId         int       `gorm:"index;default:0;not null;"`
	ContactsUserId int       `gorm:"default:0;not null;"`
	Deleted        int8      `gorm:"default:0;not null;"`
	CreateTime     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
