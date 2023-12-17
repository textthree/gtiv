package rediskey

import (
	"github.com/text3cn/goodle/kit/strkit"
	"time"
)

// 点赞关系 bitmap
const RedisKeySupports = "supports"

// 群用户禁言
func RoomBanned(roomId string) string {
	return "room:ban:" + roomId
}

// 群公告
func RoomNotice(roomId string) string {
	return "room:notice:" + roomId
}

// 群信息，在进群时需要拉群
// @return key, 缓存失效时间（单位秒）
func RoomInfo(roomId string) (string, time.Duration) {
	return "room:info:" + roomId, time.Second * time.Duration(600) // 缓存十分钟
}

// 用户在群中的信息，在进群时需要拉取
func RoomUserInfo(roomId, userId string) string {
	return "room:userinfo:" + roomId + ":" + userId
}

// 用户个人信息
func UserInfo(userId string) (string, time.Duration) {
	return "user:" + userId + ":info", time.Duration(3600*24*30) * time.Second
}

// 用户的私聊会话 key 索引
func UserMidMessageKey(toUser string) (string, time.Duration) {
	return "user:" + toUser + ":chatkeys", time.Duration(3600*24*31) * time.Second
}

// 用户的私聊会话消息信箱，为收消息的人保存信息，发消息的人本地数据库写进去了，目前不保存到服务器
func UserMidMessageList(fromUser, toUser string) (string, time.Duration) {
	return "user:" + toUser + ":chat:" + fromUser, time.Duration(3600*24*30) * time.Second
}

// 此方法在 gtiv/app/job/job/go.go 中还有一份
// 群消息，失效策略是最多保存 5000 条，如果一个群半年不活跃则这 5000 条数据失效
func RoomMessage(roomId string) (string, time.Duration) {
	if !strkit.StartWith(roomId, "room_") {
		// 增加 room_ 前缀，主要是兼容 comet 连接的 room bucket key
		roomId = "room_" + roomId
	}
	return "room:message:" + roomId, time.Duration(3600*24*30) * time.Second
}

// 申请加我为好友的列表
func ApplyContactMeList(userId string) (string, time.Duration) {
	return "user:" + userId + ":add_me", time.Duration(3600*24*30) * time.Second
}

func TimeLineSquareList() string {
	return "timeline:square"
}

// 注册验证码短信,注册时没有 uid ，用电话号码做标识
func UserRegisterVerifyCode(telStr string) (string, time.Duration) {
	return "regVerifyCode:" + telStr, time.Duration(300) * time.Second
}

// 好友关系，数据类型使用 set ，不用 bitmap
// 如果用 bitmap，用两个用户的 id 相加作为 bitIndex 会得到一个较大的值
// 那么如果用户只有 10 个好友，bitmap 也会创建够大的槽位，存一个用户也会超过 100KB
func ContactsMap(userId string) (string, time.Duration) {
	return "user:" + userId + ":contacts", time.Duration(3600*24*30) * time.Second
}
