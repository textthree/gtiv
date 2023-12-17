package entity

import (
	"time"
)

type LiveRoom struct {
	Id          uint
	UserId      uint
	RoomId      string
	Cover       string
	State       byte
	Title       string
	CreatedTime time.Time
	UpdatedTime time.Time
}
