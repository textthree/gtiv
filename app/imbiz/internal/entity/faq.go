package entity

import (
	"time"
)

type Faq struct {
	Id         uint   `gorm:"autoIncrement"`
	Title      string `gorm:"default:'';not null;"`
	Content    string `gorm:"type:text"`
	UpdateTime time.Time
}
