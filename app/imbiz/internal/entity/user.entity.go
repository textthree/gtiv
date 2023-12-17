package entity

import "time"

type User struct {
	Id              uint   `gorm:"autoIncrement"`
	Username        string `gorm:"uniqueIndex"`
	TelCode         int    `gorm:"index:idx_tel;comment:国际区号"`
	Tel             string `gorm:"index:idx_tel;comment:手机号"`
	Password        string
	Role            int8 `gorm:"default:10;comment:用户角色 10.普通用户 20.系统管理员，啥权限都有 100.开发人员"`
	Nickname        string
	Birthday        string
	Gender          int8 `gorm:"default:0;not null;comment:性别 0.保密 1.男 2.女"`
	Avatar          string
	LastLoginTime   int       `gorm:"comment:最后登录时间"`
	ContactsVersion int       `gorm:"default:0;not null"`
	Intro           string    `gorm:"type:text;comment:个人简介"`
	FansNum         int       `gorm:"comment:粉丝数;default:0;not null"`
	FollowNum       int       `gorm:"comment:关注的人数量;default:0;not null"`
	CollectNum      int       `gorm:"comment:我发布的视频被收藏的次数;default:0;not null"`
	SupportNum      int       `gorm:"comment:我收到的点赞次数;default:0;not null"`
	CollectVideoNum int       `gorm:"comment:我收藏的视频数;default:0;not null"`
	CreateVideoNum  int       `gorm:"comment:我创建的视频数;default:0;not null"`
	AccessTime      time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	CreateTime      time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdateTime      time.Time `gorm:"default:null"`
}
