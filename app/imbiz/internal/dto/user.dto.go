package dto

// 注册
type UserRegisterReq struct {
	Tel        int    // 电话
	TelCode    int    // 国际电话区号
	VerifyCode int    // 验证码
	Password   string // 密码
}
type UserRegisterRes struct {
	BaseRes
	Token         string
	LastLoginTime int // 最后登录时间
	Uid           string
}

// 登录
type UserLoginReq struct {
	Tel      string
	TelCode  int
	Username string
	Password string
}
type UserLoginRes struct {
	BaseRes
	Userinfo Userinfo
}

type Userinfo struct {
	UserId          int
	UserRole        int    // 用户角色
	Nickname        string // 昵称
	Username        string // 登录用户名
	Avatar          string // 头像
	Gender          int    // 性别
	FansNum         int
	CollectNum      int
	SupportNum      int
	CreateVideoNum  int
	FollowNum       int
	ContactsVersion int // 通讯录版本号
	Token           string
	LastLoginTime   int // 最后登录时间
}

// 获取用户信息
type UserinfoReq struct {
	UserId string
}
type UserinfoRes struct {
	BaseRes
	UserId          string
	UserRole        int    // 用户角色
	Nickname        string // 昵称
	Username        string // 登录用户名
	Avatar          string // 头像
	Gender          int    // 性别
	Birthday        string // 生日
	LastLoginTime   int
	ContactsVersion int // 通讯录版本号
	FansNum         int64
	FollowNum       int64
	SupportNum      int64
	CollectVideoNum int64
	CreateVideoNum  int64
}

// 修改用户信息
type UpdateUserinfoReq struct {
	Field string // 要修改的字段
	Value string // 修改的值
}
type UpdateUserinfoRes struct {
	BaseRes
}

// 查找用户
type SearchUserReq struct {
	Username string // 国际区号+电话
}
type SearchUserRes struct {
	BaseRes
	UserId   string
	Username string
	Nickname string
	Avatar   string
}

// 我关注的用户列表
type SubscribeUserListReq struct {
	Page int // 第几页
	Rows int // 每页显示多少条
}
type SubscribeUserList struct {
	UserId   string // 被关注者用户 ID
	Nickname string // 被关注者昵称
	Avatar   string // 被关注者头像
}
type SubscribeUserListRes struct {
	BaseRes
	List []SubscribeUserList // 我关注的用户列表
}

// 判断是否关注某人
type IsSubscribeUserReq struct {
	UserId string // 被判断是否关注者的用户 ID
}
type IsSubscribeUserRes struct {
	BaseRes
	IsSubscribe bool // 是否已关注
}
