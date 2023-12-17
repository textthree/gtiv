package entity

type Country struct {
	Id             uint   `gorm:"autoIncrement"`
	Name           string `gorm:"default:'';not null;"`
	NameEn         string `gorm:"default:'';not null;"`
	CountryCode    int    `gorm:"comment:国际域名缩写"`
	TelCode        int    `gorm:"comment:国际区号"`
	TimeDifference int    `gorm:"comment:与中国时差"`
	Open           int8
}
