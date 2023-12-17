package caches

import (
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/providers/goodlog"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"github.com/text3cn/goodle/providers/redis"
	"gtiv/app/imbiz/internal/caches/rediskey"
	"gtiv/app/imbiz/internal/dto"
)

// 重建用户信息缓存
func RebuildUserinfoCache(ctx *httpserver.Context, userId string) {
	conn := redis.Instance().Conn()
	key, expire := rediskey.UserInfo(userId)
	row := queryUser(ctx, userId)
	res := conn.Set(ctx.BaseContext(), key, strkit.JsonEncode(row), expire)
	if res.Err() != nil {
		goodlog.Error(res.Err())
	}
}

// 去数据库查用户信息
func queryUser(ctx *httpserver.Context, userId string) *dto.Userinfo {
	db := orm.GetDB()
	sql := "SELECT id as UserId, role as UserRole, version, contacts_version, username, nickname, avatar," +
		" fans_num, follow_num, support_num, create_video_num, collect_video_num, " +
		" last_login_time, gender, birthday " +
		" FROM user WHERE id = " + userId
	var userinfo dto.Userinfo
	db.Raw(sql).Scan(&userinfo)
	return &userinfo
}

// userinfo 是查询频次较高的，建议双写
func GetUserinfo(ctx *httpserver.Context, userId string) (userinfo *dto.Userinfo) {
	conn := redis.Instance().Conn()
	key, expire := rediskey.UserInfo(userId)
	cache := conn.Get(ctx.BaseContext(), key)
	if cache.Val() != "" && cache.Val() != "null" {
		strkit.JsonDecode(cache.Val(), &userinfo)
	} else {
		userinfo = queryUser(ctx, userId)
		res := conn.Set(ctx.BaseContext(), key, strkit.JsonEncode(userinfo), expire)
		if res.Err() != nil {
			goodlog.Error(res.Err())
		}
	}

	return
}
