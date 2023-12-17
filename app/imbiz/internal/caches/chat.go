package caches

import (
	"context"
	"github.com/spf13/cast"
	"github.com/text3cn/goodle/kit/timekit"
	"github.com/text3cn/goodle/providers/redis"
	"gtiv/app/imbiz/internal/caches/rediskey"
)

// 获取用户是否禁言，同时判断禁言是否到期，如果到期则删除这条记录
func CheckRoomBannedExpire(roomId, userId string) bool {
	RoomBannedKey := rediskey.RoomBanned(roomId)
	conn := redis.Instance().Conn()
	reply := conn.Do(context.Background(), "HGET", RoomBannedKey, userId)
	now := timekit.NowTimestamp()
	res, _ := reply.Result()
	if now-cast.ToInt(res) > 0 {
		// 删除条目
		conn.Do(context.Background(), "HDEL", RoomBannedKey, userId)
		return false
	}
	return true
}
