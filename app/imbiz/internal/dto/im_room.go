package dto

import "time"

// 进群获取群信息
type RoomInfoReq struct {
	RoomId string
}
type RoomInfoRes struct {
	BaseRes
	BannedAll  bool   // 全员禁言结束时间
	BannedUser bool   // 用户个人禁言结束时间
	Id         string // 房间 id
	Video      string // 群视频
	Avatar     string // 群头像
	MemberNum  int64  // 群人数
	RoomName   string // 群名称
	RoomType   int64  // 群类型 1.官方群 2.用户群
	ShowVideo  int64  // 是否展示视频 1.是 0.否
	IsMember   bool   // 用户是否属于当前群成员
	Role       int64  // 用户在此群的角色：0.游客 10.群成员 20.管理员 30.群主
}

// 获取指定用户在某个群中的信息
type RoomUserInfoReq struct {
	RoomId string
	UserId string
}
type RoomUserInfoRes struct {
	BaseRes
	Role int64 // 用户在此群的角色：0.游客 10.群成员 20.管理员 30.群主
}

// 拉取最近群消息
type RoomMessageListReq struct {
	Room_id string
	Page    int // 第几页
}
type RoomMessageListRes struct {
	BaseRes
	Time int
	Data []RoomMessageItem
}
type RoomMessageListRet struct {
	BaseRes
	Data interface{}
	Time int
}
type RoomMessageItem struct {
	RoomId      int
	FromUser    int
	ToUsers     string
	Message     string
	Type        int
	UserVersion int    `v:"required"` // 消息发送者的用户信息版本号
	MsgId       string `v:"required"` // 消息 id
}

// 群成员列表
type RoomMemberListReq struct {
	RoomId   string // 群 ID
	Nickname string // 按名称搜索
	Page     int    // 第几页
	Rows     int    // 每页显示多少条
}
type RoomMemberListRes struct {
	BaseRes
	List []RoomMemberListItem
}
type RoomMemberListItem struct {
	Avatar     string
	Nickname   string
	UserId     string
	Role       int // 用户在此群的角色 10.普通用户 20.群普通管理员 30.群主
	CreateTime time.Time
}

// 群成员信息
type RoomMemberInfoReq struct {
	RoomId string // 群 ID
	UserId string // 用户 id
}
type RoomMemberInfoRes struct {
	BaseRes
	Avatar     string
	Nickname   string
	CreateTime int
	Role       int
}

// 修改群名称
type RoomModifyNameReq struct {
	RoomId string // 群 ID
	Name   string // 群名称
}
type RoomModifyNameRes struct {
	BaseRes
}

// 修改群头像
type RoomModifyAvatarReq struct {
	RoomId string
	Avatar string // 头像 url
}
type RoomModifyAvatarRes struct {
	BaseRes
}

// 编辑群公告
type RoomModifyNoticeReq struct {
	RoomId  string // 群 ID
	Content string // 群公告内容
}
type RoomModifyNoticeRes struct {
	BaseRes
}

// 查看群公告
type RoomGetNoticeReq struct {
	RoomId string // 群 ID
}
type RoomGetNoticeRes struct {
	BaseRes
	Content string // 群公告内容
}

// 禁言
type RoomBannedToPostReq struct {
	RoomId string
	UserId string // 被禁的用户 ID
	Type   int    // 禁言类型 1、10分钟 2、1小时 3、12小时 4、1天 5、永久禁言
}
type RoomBannedToPostRes struct {
	BaseRes
}

// 解除禁言
type RoomRelieveBannedToPostReq struct {
	RoomId string // 群 ID
	UserId string // 被禁的用户 ID，为 0 代表全体禁言
}
type RoomRelieveBannedToPostRes struct {
	BaseRes
}

// 禁言列表
type RoomBannedListReq struct {
	RoomId string
}
type RoomBannedListRes struct {
	BaseRes
	List []RoomBannedList
}
type RoomBannedList struct {
	UserId   string
	Expire   int    // 禁言结束时间
	Nickname string // 昵称
	Avatar   string // 头像
}

// 群管理员列表
type RoomAdminListReq struct {
	RoomId string
}
type RoomAdminListRes struct {
	BaseRes
	List []RoomAdminList
}
type RoomAdminList struct {
	UserId   string
	Nickname string
	Avatar   string
	Role     int // 角色：10.普通用户 20.群普通管理员 30.群主
}

// 设置管理员
type RoomSetAdminReq struct {
	RoomId  string
	UserIds string // 最终成为管理员的那些用户的 id，多个逗号隔开
}
type RoomSetAdminRes struct {
	BaseRes
}

// 常见问答
type RoomFaqListReq struct {
}
type RoomFaqItem struct {
	Title   string
	Content string
}
type RoomFaqListRes struct {
	BaseRes
	List []RoomFaqItem
}

// 创建聊天室/邀请成员
type InviteMemberReq struct {
	RoomId  string `dc:"如果值非空则代表邀请邀请，否则为新创建群"`
	UserIds string `dc:"邀请的成员，多个逗号隔开"`
	Message string `dc:"发一条谁邀请了谁的消息到群中"`
}
type InviteMemberRes struct {
	BaseRes
	RoomId   string `v:"required"`
	Message  string `v:"required"`
	RoomName string `v:"required"`
}

// 拉取所有在群中的用户的 id
type RoomMemberIdsReq struct {
	RoomId string
}
type RoomMemberIdsRes struct {
	BaseRes
	List []int
}

// 移除群成员
type RemoveMemberReq struct {
	RoomId string `v:"required" dc:"群 ID"`
	UserId string `v:"required" dc:"用户 id"`
}
type RemoveMemberRes struct {
	BaseRes
}

// 退出群聊
type QuitRoomReq struct {
	RoomId string
}
type QuitRoomRes struct {
	BaseRes
}

// 解散群
type DissolveRoomReq struct {
	RoomId string
}
type DissolveRoomRes struct {
	BaseRes
}

// 用户的群列表
type RoomListReq struct {
}
type RoomList struct {
	RoomId      string
	RoomName    string
	Avatar      string
	MemberNum   int
	MemberLimit int
}
type RoomListRes struct {
	BaseRes
	List []RoomList
}
