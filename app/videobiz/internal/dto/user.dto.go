package dto

type RegisterRes struct {
	BaseRes
	Userinfo Userinfo
}

type Userinfo struct {
	UserId          int    `v:"required"`
	UserRole        int    `v:"required" dc:"用户角色"`
	Nickname        string `v:"required" dc:"昵称"`
	Username        string `v:"required" dc:"登录用户名"`
	Avatar          string `v:"required" dc:"头像"`
	Gender          int    `v:"required" dc:"性别"`
	Version         int    `v:"required" dc:"用户信息版本号，前端修改用户信息时改变"`
	ContactsVersion int    `v:"required" dc:"通讯录版本号"`
	Token           string `v:"required"`
	LastLoginTime   int    `v:"required" dc:"最后登录时间"`
}

type VideoMasterInfoRes struct {
	BaseRes
	Nickname       string
	Avatar         string
	Username       string
	SupportNum     int
	FollowNum      int
	FansNum        int
	CreateVideoNum int
	Intro          string
	IsFollow       bool
	IsFriend       bool
}

type MyUserListItem struct {
	Id         int
	Nickname   string
	Avatar     string
	CreateTime TimeToUnix
}
type MyUserListRes struct {
	BaseRes
	List []MyUserListItem
}
