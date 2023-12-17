package im

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/text3cn/goodle/kit/cryptokit"
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/providers/goodlog"
	"github.com/text3cn/goodle/providers/orm"
	"gtiv/app/imbiz/boot"
	"gtiv/app/imbiz/im/model"
	"gtiv/kit/impkg/protocol"
	"time"
)

// Connect connected a conn.
// App 与 Comet 建立连接后发送授权，Comet 通过 RPC 请求 logic 来这鉴权
func (l *Logic) Connect(c context.Context, server, cookie string, token []byte) (
	mid int64, key, roomID string, accepts []int32, hb int64, err error) {
	// 客户端携带过来的参数
	var params struct {
		Key       string `json:"key"` // 连接 key，现在客户端没有传递 key 过来
		Platform  string `json:"platform"`
		UserToken string `json:userToken`
	}
	if err = json.Unmarshal(token, &params); err != nil {
		goodlog.Error("json.Unmarshal error", token, err)
		return
	}
	userId := cryptokit.DynamicDecrypt(boot.TokenKey, params.UserToken)
	mid = strkit.ParseInt64(userId)
	roomID = ""
	sql := "SELECT room_id FROM user_room WHERE user_id = ?"
	roomIds := []int32{}
	orm.GetDB().Raw(sql, userId).Scan(&roomIds)
	if len(roomIds) > 0 {
		for _, roomId := range roomIds {
			// 如果屏蔽消息的群不接收，可以在这里处理
			// 100 以内的 operation 为系统保留，所以给 room_id 加上 100
			accepts = append(accepts, roomId+100)
		}
	}
	accepts = append(accepts, protocol.OpMidMsg)
	// 连接多久过期，单位秒，客户端是 30 秒发送一次心跳
	// 如果 300 秒还没有收到客户端发送来心跳则 comet 主动断开连接
	hb = 300
	if key = params.Key; key == "" {
		// 192.168.1.12 也运行者一个 logic，通过 killall boom_alpha 停止
		// 注意这个 key 不能使用固定值，要拼接 uuid，如果使用固定值客户端无法断线重连，
		// 因为重连时会先断开旧连接然后建立新连接，而断开需要四次挥手需要一定时间，不同机型上这两个操作并没有保证时序，
		// 客户端这样做是没有问题的，但是服务端使用了 key 来管理连接，所以如果使用相同的 key 就可能先建立了连接然后又被断开了，
		// 因此使用了 uuid 让每次连接的 key 不同，那么多次连接就不会相互影响
		key = params.Platform + "-" + strkit.Tostring(mid) + "-" + uuid.New().String()
	}
	if err = l.dao.AddMapping(c, mid, key, server); err != nil {
		goodlog.Error("l.dao.AddMapping error", mid, key, server, err)
	}
	goodlog.Info("conn connected, key:", key, server, mid)
	return
}

// Disconnect 断开连接
func (l *Logic) Disconnect(c context.Context, mid int64, key, server string) (has bool, err error) {
	if has, err = l.dao.DelMapping(c, mid, key, server); err != nil {
		goodlog.Errorf("l.dao.DelMapping(%d,%s) error(%v)", mid, key, server)
		return
	}
	goodlog.Infof("conn disconnected key:%s server:%s mid:%d", key, server, mid)
	return
}

// Heartbeat heartbeat a conn.
func (l *Logic) Heartbeat(c context.Context, mid int64, key, server string) (err error) {
	has, err := l.dao.ExpireMapping(c, mid, key)
	if err != nil {
		goodlog.Errorf("l.dao.ExpireMapping(%d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if !has {
		if err = l.dao.AddMapping(c, mid, key, server); err != nil {
			goodlog.Errorf("l.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	goodlog.Infof("conn heartbeat key:%s server:%s mid:%d", key, server, mid)
	return
}

// RenewOnline renew a server online.
func (l *Logic) RenewOnline(c context.Context, server string, roomCount map[string]int32) (map[string]int32, error) {
	online := &model.Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := l.dao.AddServerOnline(context.Background(), server, online); err != nil {
		return nil, err
	}
	return l.roomCount, nil
}
