package services

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/text3cn/goodle/kit/arrkit"
	"github.com/text3cn/goodle/kit/castkit"
	"github.com/text3cn/goodle/kit/gokit"
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/kit/timekit"
	"github.com/text3cn/goodle/providers/goodlog"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"github.com/text3cn/goodle/providers/redis"
	"gorm.io/gorm"
	"gtiv/app/imbiz/im"
	"gtiv/app/imbiz/internal/caches"
	"gtiv/app/imbiz/internal/caches/rediskey"
	"gtiv/app/imbiz/internal/constants"
	"gtiv/app/imbiz/internal/dto"
	"gtiv/kit/impkg/protocol"
)

type chat struct {
	db       *gorm.DB
	uid      *castkit.GoodleVal
	ctx      *httpserver.Context
	userinfo *dto.Userinfo
}

var chatInstance *chat

func Chat(ctx *httpserver.Context) *chat {
	if chatInstance != nil {
		return chatInstance
	}
	uid := ctx.GetVal("uid")
	userinfo := caches.GetUserinfo(ctx, uid.ToString())
	return &chat{
		db:       orm.GetDB(),
		uid:      uid,
		ctx:      ctx,
		userinfo: userinfo,
	}
}

// 同意或拒绝添加联系人
func (this chat) PushMid(req *dto.PushMidReq) (serverMsgId string, time int, err error) {
	Op := protocol.OpMidMsg // 这个 op 对应客户端 AbstractBlockingClient.java 中的 operation
	time = timekit.Millisecond()
	serverMsgId = uuid.New().String()
	midMessage := dto.PushMidMessageDto{
		ChatType:    1,
		MsgType:     req.Type,
		ToUsers:     req.ToUsers,
		Content:     req.Message,
		ServerMsgId: serverMsgId,
		FromUser:    this.uid.ToString(),
		Time:        time,
		BadgeNum:    1,
	}
	// TODO 判断是否好友关系
	mids := []int64{strkit.ParseInt64(req.ToUsers)}
	msgBodyJson, err := json.Marshal(midMessage)
	// 申请加好友
	if req.Type == constants.MessageTypeSayHello {
		key, expire := rediskey.ApplyContactMeList(req.ToUsers)
		conn := redis.Instance().Conn()
		if result := conn.Do(this.ctx, "HSET", key, midMessage.FromUser, msgBodyJson); result.Err() != nil {
			goodlog.Error("[申请加好友存储 Redis 失败]" + err.Error())
		}
		conn.Do(this.ctx, "EXPIRE", key, expire)
	} else if req.Type == constants.ApplyOne2OneVideoCall {
		msgBodyJson, err = json.Marshal(midMessage)
		im.ImLogic.PushToMids(this.ctx, Op, mids, msgBodyJson)
		return
	} else {
		dontSave := []int{
			constants.MessageNewRoom,
			constants.MessageDeleteContacts,
			constants.OtherPlaceSignIn,
			constants.AcceptCall,
		}
		if !arrkit.InArray(req.Type, dontSave) {
			// 先备份到 redis 在发送消息，确保消息的不丢失
			// TODO 这个操作应该放在 kafka 中，不然要消息队干嘛
			this.saveMidMsg(midMessage, msgBodyJson)
		}
	}
	im.ImLogic.PushToMids(this.ctx, Op, mids, msgBodyJson)
	return
}

// 保存私聊消息
func (this chat) saveMidMsg(midMessage dto.PushMidMessageDto, messageBody []byte) {
	conn := redis.Instance().Conn()
	// 保存索引
	key, expire := rediskey.UserMidMessageKey(midMessage.ToUsers)
	conn.Do(this.ctx, "SADD", key, midMessage.FromUser)
	conn.Do(this.ctx, "EXPIRE", key, expire)
	// 保存内容
	key, expire = rediskey.UserMidMessageList(midMessage.FromUser, midMessage.ToUsers)
	conn.Do(this.ctx, "ZADD", key, midMessage.Time, messageBody)
	conn.Do(this.ctx, "EXPIRE", key, expire)
}

// 发送群消息，默认好像广播做了节流，1 秒内同一个用户发送多条消息的行为会被阻止
// TODO 加提示：你说话的速度太快了，请稍作休息。我们的人力物理和投入的时间，目前来说没办法和经过十几年发展的 IM 软件对标
// goim 原来是单房间设计，切换房间则重新建立长连接订阅新的房间，适用于直播进房间场景
func (this chat) PushRoom(req *dto.PushRoomReq) (serverMsgId string, time int, err error) {
	conn := redis.Instance().Conn()
	roomId := "room_" + req.RoomId
	// 对于群消息，只要用户订阅的 accepts 列表中包含这个 operation 都能收到这条消息
	// 这个 operation 在客户端 AbstractBlockingClient.java 中的 operation 用
	// 前 100 的 operation 为 comet 占用，所以房间 id 都加上 100
	Op := int32(strkit.ParseInt(req.RoomId)) + 100
	time = timekit.Millisecond()
	serverMsgId = uuid.New().String()
	data := dto.PushRoomMessageDto{
		ChatType:    2,
		MsgType:     req.Type,
		FromUser:    this.uid.ToString(),
		ToUsers:     req.ToUsers,
		Content:     req.Message,
		ServerMsgId: serverMsgId,
		RoomId:      req.RoomId,
		Time:        time,
	}
	msgBodyJson, err := json.Marshal(data)
	if err != nil {
		goodlog.Error("push room message append time marshal error.")
	}
	// 判断是否群成员、是否已禁言、是否已被挤下线
	roomInfo, _ := Room(this.ctx).RoomInfo(req.RoomId)
	isMember, role, _ := Room(this.ctx).RoomUserinfo(req.RoomId)
	if !isMember {
		err = errors.New("You are not room member, can not send message.")
		return
	}
	if (roomInfo.BannedUser || roomInfo.BannedAll) && int(role) == constants.RoomUserRole.Normal {
		err = errors.New("Banned to post now.")
		return
	}
	// redis 缓存每个群的最近 5000 条消息
	go gokit.SafeGo(func() {
		key, expire := rediskey.RoomMessage(roomId)
		conn.Do(this.ctx, "RPUSH", key, msgBodyJson)
		if result := conn.Do(this.ctx, "LTRIM", key, -5000, -1); result.Err() != nil {
			goodlog.Error("Room message LTRIM error" + err.Error())
		}
		conn.Do(this.ctx, "EXPIRE", key, expire)
	})
	im.ImLogic.PushAll(this.ctx, Op, 0, msgBodyJson)
	return
}

func (this chat) AddMeList() (ret []dto.AddMeListItem, err error) {
	conn := redis.Instance().Conn()
	key, _ := rediskey.ApplyContactMeList(this.uid.ToString())
	r := conn.Do(this.ctx, "HGETALL", key).Val()
	list, _ := r.(map[interface{}]interface{})
	for _, v := range list {
		mp, _ := strkit.Json_decode_map(v.(string))
		sql := "SELECT avatar, nickname, username, gender FROM user WHERE id = ?"
		var avatar, nickname, username string
		var gender int8
		this.db.Raw(sql, mp["FromUser"]).Row().Scan(&avatar, &nickname, &username, &gender)
		FromUser, _ := mp["FromUser"].(string)
		Msg, _ := mp["Content"].(string)
		Time, _ := mp["Time"].(float64)
		item := dto.AddMeListItem{
			UserId:   strkit.ParseInt(FromUser),
			Avatar:   avatar,
			Nick:     nickname,
			Msg:      Msg,
			Time:     int(Time / 1000),
			Gender:   gender,
			Username: username,
		}
		ret = append(ret, item)
	}
	return
}

// 获取未同步的聊天记录
// fromUserId 获取与谁的聊天记录
// lastMessageTime 客户端数据库与该用户聊天的最后一条消息时间
// 赶时间，不分页，简单粗暴一次性取出所有吧，谁不在线还大量给他发消息，打死他
func (this chat) ChatRecord(lastMessageTime int) (ret []dto.UserMessageListItem, err error) {
	conn := redis.Instance().Conn()
	// 先查出索引然后挨个同步
	indexKey, _ := rediskey.UserMidMessageKey(this.uid.ToString())
	keys := conn.Do(this.ctx, "SMEMBERS", indexKey).Val()
	keysArr, _ := keys.([]interface{})
	if len(keysArr) == 0 {
		return
	}
	for _, v := range keysArr {
		fromUser, _ := v.(string)
		key, _ := rediskey.UserMidMessageList(fromUser, this.uid.ToString())
		//echo.P("同步" + key + "tiem")
		// lastMessageTime 要加 1，不加 1 是大于等于
		r := conn.Do(this.ctx, "ZRANGEBYSCORE", key, lastMessageTime+1, "+inf").Val()
		list, _ := r.([]interface{})
		if len(list) > 0 {
			for _, vv := range list {
				data, _ := strkit.Json_decode_map(vv.(string))
				msgId, _ := data["MsgId"].(string)
				msgType, _ := data["MsgType"].(float64)
				fromUser, _ := data["FromUser"].(string)
				content, _ := data["Content"].(string)
				time, _ := data["Time"].(float64)
				item := dto.UserMessageListItem{
					MsgId:    msgId,
					MsgType:  int(msgType),
					FromUser: strkit.ParseInt(fromUser),
					Content:  content,
					Time:     int64(time),
				}
				ret = append(ret, item)
			}
		}
		// 不管查不查得到数据，来同步过一次就直接清空
		// 这样其实不保险，应该是前端前部同步完成后再发送一个回执再清空，
		// 如果同步中断可支持继续传最新时间同步，直到收到客户端回执才去清空 redis
		conn.Do(this.ctx, "DEL", key)
		conn.Do(this.ctx, "SREM", indexKey, fromUser)
	}
	return
}

// 获取群聊记录
func (this chat) RoomMsg(roomId string, lastMessageTime int64) (ret []dto.RoomMsgListItem, err error) {
	conn := redis.Instance().Conn()
	key, _ := rediskey.RoomMessage(roomId)
	// 从倒数第 1 条取到倒数第 30 条，往回遍历取 lastMessageTime 之后的消息，超出 30 条算球
	r := conn.Do(this.ctx, "LRANGE", key, -30, -1).Val()
	list := r.([]interface{})
	length := len(list)
	if length > 0 {
		for i := length - 1; i >= 0; i-- {
			mp, _ := strkit.Json_decode_map(list[i].(string))
			msgId, _ := mp["MsgId"].(string)
			msgType, _ := mp["MsgType"].(float64)
			time, _ := mp["Time"].(float64)
			fromUser, _ := mp["FromUser"].(string)
			msg, _ := mp["Content"].(string)
			//fmt.Println(lastMessageTime, "|", int64(time))
			if lastMessageTime >= int64(time) {
				break
			}
			item := dto.RoomMsgListItem{
				MsgId:    msgId,
				MsgType:  int64(msgType),
				FromUser: strkit.ParseInt(fromUser),
				Time:     int64(time),
				Content:  msg,
			}
			ret = append(ret, item)
		}
	}
	return
}
