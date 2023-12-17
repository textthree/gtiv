package constants

const (
	// 聊天资源，七天过期
	OBS_BUCKET_CHAT = iota

	// 其他资源，不过期
	OBS_BUCKET_IM

	// 视频资源，会截缩图
	OBS_BUCKET_VIDEO
)

// 聊天室视频类型
const (
	// 系统默认视频
	ROOM_VIDEO_TYPE_SYSTEM = 1
	// 用户上传的视频
	ROOM_VIDEO_TYPE_USER = 2
)

// 会话类型
const (
	// 私聊
	CHAT_TYPE_PRIVATE = 1

	// 群聊
	CHAT_TYPE_ROOM = 2
)

// 消息类型
const (
	MessageTypeText       = 0
	MessageTypeImage      = 1
	MessageTypeAudio      = 2
	MessageTypeVideo      = 3
	MessageTypeBanToPost  = 4  // 禁言
	MessageTypeRelieveBan = 5  // 解除禁言
	MessageTypeRepeal     = 6  // 撤回消息
	MessageTypeSayHello   = 7  // 打招呼消息
	MessageNewChat        = 8  // 好友验证通过
	MessageNewRoom        = 9  // 被拉进新群
	ApplyOne2OneVideoCall = 10 // 请求一对一视频通话
	MessageDeleteContacts = 13 // 删除联系人
	OtherPlaceSignIn      = 16 // 其他地方登录
	AcceptCall            = 17 // 对方接听了我的通话请求
	InviteMember          = 20
)
