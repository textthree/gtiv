package dto

// IM 消息格式
type ImMessage struct {
	RoomId      string
	FromUser    string
	ToUsers     string // 如果有值代表艾特别人
	Message     string
	UserVersion int    // 消息发送者的用户信息版本，后端这里主要用于推送禁言/解除禁言类的管理员消息，传 0 即可
	MsgId       string // 消息 id，使用客户端生成 guid
	Type        int    // 消息类型
	Time        int    // 非必填
}

// 私聊消息包
type PushMidMessageDto struct {
	ChatType    int    // 聊天类型 1.私聊 2.群聊
	MsgType     int    // 消息类型
	FromUser    string // 发送者（前端不用传，后端会覆盖）
	ToUsers     string // 接收者
	Content     string // 消息内容
	ServerMsgId string // 消息 id
	Time        int    // 毫秒时间戳（前端不用传，后端会覆盖）
	BadgeNum    int64  // 前端不用传
}

// 群聊消息包
type PushRoomMessageDto struct {
	ChatType    int // 聊天类型 1.私聊 2.群聊
	MsgType     int
	RoomId      string
	FromUser    string
	ToUsers     string
	Content     string // 原生直接原样渲染的消息
	ServerMsgId string
	Time        int // 毫秒时间戳
}

type PushMidReq struct {
	Type    int    // 消息类型
	ToUsers string // 接收者
	Message string // 消息内容
}
type PushMidRes struct {
	BaseRes
	Time        int // 消息时间
	ServerMsgId string
}

// 发送群消息
type PushRoomReq struct {
	Type    int    // 消息类型
	ToUsers string // 接收者
	Message string // 消息文本，用于客户端原样输出
	RoomId  string // 房间 id
}
type PushRoomRes struct {
	BaseRes
	Time        int // 消息时间
	ServerMsgId string
}

// 判断用户是否在线
type CheckOnlineReq struct {
	User string // 目标用户，不传则检测自己是否在线
}
type CheckOnlineRes struct {
	BaseRes
	Online bool // 是否在线
}

// 获取申请加我为好友的人
type AddMeListReq struct {
}

type AddMeListItem struct {
	UserId   int
	Avatar   string
	Nick     string
	Msg      string
	Time     int
	Username string
	Gender   int8
}
type AddMeListRes struct {
	BaseRes
	List []AddMeListItem
}

type ChatItem struct {
	Id              string // 用户 id 或 群 id
	ChatType        int    // 会话类型：1.私聊 2.群聊
	BadgeNum        int    // 从上次收到最后一条消息时间起，未读消息数
	LastMessageTime int    // 最后一条消息时间
	LastMessageDesc string // 最后一条消息内容摘要
	UserVersion     int    // 最后一条消息时的用户信息版本号
	RoomMemberNum   int
}

// 获取未同步的聊天记录
type ChatRecordReq struct {
	LastMessageTime int // 客户端数据库与该用户聊天的最后一条消息时间
}
type UserMessageListItem struct {
	MsgId    string
	MsgType  int
	FromUser int
	Content  string
	Time     int64
}
type ChatRecordRes struct {
	BaseRes
	List []UserMessageListItem
}

// 获取指定群聊天记录
type RoomMsgListItem struct {
	MsgId    string
	MsgType  int64
	FromUser int
	Time     int64
	Content  string
}

// 获取未同步的聊天记录
type RoomMsgReq struct {
	RoomId          string
	LastMessageTime int // 客户端保存的该群最后一条聊天记录
}
type RoomMsgRes struct {
	BaseRes
	List []RoomMsgListItem
}
