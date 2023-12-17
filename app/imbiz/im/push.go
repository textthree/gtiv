package im

import (
	"context"
	"github.com/text3cn/goodle/providers/goodlog"
)

// PushKeys push a message by keys.
func (l *Logic) PushKeys(c context.Context, op int32, keys []string, msg []byte) (err error) {
	servers, err := l.dao.ServersByKeys(c, keys)
	if err != nil {
		return
	}
	pushKeys := make(map[string][]string)
	for i, key := range keys {
		server := servers[i]
		if server != "" && key != "" {
			pushKeys[server] = append(pushKeys[server], key)
		}
	}
	for server := range pushKeys {
		if err = l.dao.PushMsg(c, op, server, pushKeys[server], msg); err != nil {
			return
		}
	}
	return
}

// PushToMids push a message by mid.
func (l *Logic) PushToMids(c context.Context, op int32, mids []int64, msg []byte) (err error) {

	keyServers, olMids, err := l.dao.KeysByMids(c, mids)

	if err != nil {
		return
	}
	if len(olMids) == 0 {
		goodlog.Info("用户不在线")
		//return
	}
	keys := make(map[string][]string)
	for key, server := range keyServers {
		if key == "" || server == "" {
			goodlog.Errorf("push key:%s server:%s is empty", key, server)
			continue
		}
		keys[server] = append(keys[server], key)
	}
	for server, keys := range keys {
		if err = l.dao.PushMsg(c, op, server, keys, msg); err != nil {
			return
		}
	}
	return
}

// PushRoom push a message by room.
func (l *Logic) PushRoom(c context.Context, op int32, room string, msg []byte) (err error) {
	return l.dao.BroadcastRoomMsg(c, op, room, msg)
}

// PushAll push a message to all.
func (l *Logic) PushAll(c context.Context, op, speed int32, msg []byte) (err error) {
	return l.dao.BroadcastMsg(c, op, speed, msg)
}

// 检测目标用户是否在线
func (l *Logic) CheckOnline(c context.Context, mids []int64) bool {
	_, olMids, _ := l.dao.KeysByMids(c, mids)
	return len(olMids) > 0
}
