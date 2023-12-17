package user

import (
	"github.com/spf13/cast"
	"github.com/text3cn/goodle/kit/cryptokit"
	"github.com/text3cn/goodle/kit/mathkit"
	"github.com/text3cn/goodle/kit/timekit"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"gtiv/app/videobiz/internal/dto"
)

// 使用账号密码注册
func RegisterByUsername(ctx *httpserver.Context) {
	ret := dto.RegisterRes{}
	username, _ := ctx.Req.FormString("username")
	password, _ := ctx.Req.FormString("password")
	if orm.CheckEmptyId("SELECT id FROM user WHERE username = ?", username) {
		ret.ApiCode = 1
		ret.ApiMessage = "User name already exists" + password
		ctx.Resp.Json(ret)
		return
	}
	lastLoginTime := timekit.NowTimestamp()
	sql := "INSERT INTO user(username, nickname, password, last_login_time) VALUES(?, ?, ?, ?)"
	nickname := "User" + cast.ToString(mathkit.Rand(1000, 9999))
	db := orm.GetDB()
	db.Exec(sql, username, nickname, cryptokit.Md5(password+"t3im"), lastLoginTime)
	id := 0
	db.Raw("SELECT id FROM user WHERE username = ?", username).Scan(&id)
	token := cryptokit.DynamicEncrypt(ctx.Config.Get("app.token_key").ToString(), cast.ToString(id))
	ret.Userinfo = dto.Userinfo{
		UserId:        id,
		Nickname:      nickname,
		Username:      username,
		LastLoginTime: lastLoginTime,
		Token:         token,
	}
	ctx.Resp.Json(ret)
	return
}

// 获取播主信息，并判断是否关注、是否是好友关系
func VideoMasterInfo(ctx *httpserver.Context) {
	ret := dto.VideoMasterInfoRes{}
	myUid := ctx.GetVal("userId").ToInt()
	masterUid := ctx.Req.GetInt("userId")
	db := orm.GetDB()

	sql := "SELECT nickname, avatar, username, intro, " +
		" support_num, follow_num, fans_num, create_video_num " +
		" FROM user WHERE id = ?"
	db.Raw(sql, masterUid).Scan(&ret)

	// 关注状态、是否好友
	if myUid == masterUid {
		ret.IsFollow = true
		ret.IsFriend = true
	} else if myUid > 0 {
		// 关注状态
		id := 0
		sql = "SELECT id FROM follow WHERE fans = ? AND master = ? LIMIT 1"
		db.Raw(sql, myUid, masterUid).Scan(&id)
		if id > 0 {
			ret.IsFollow = true
		}
		// 是否好友
		id = 0
		sql = "SELECT id FROM contacts WHERE user_id = ? AND contacts_user_id = ? LIMIT 1"
		db.Raw(sql, myUid, masterUid).Scan(&id)
		if id > 0 {
			ret.IsFriend = true
		}
	}
	ctx.Resp.Json(ret)
}

// 获取用户发布的视频
func CreationVideos(ctx *httpserver.Context) {
	db := orm.GetDB()
	masterUid := ctx.Req.GetInt("userId")
	page := ctx.Req.GetInt("page")
	rows := 12
	skip := (page - 1) * rows
	sql := "SELECT id video_id, title, cover, support_num FROM videos WHERE user_id = ? LIMIT ?, ?"
	list := []dto.UserHomeVideoItem{}
	db.Raw(sql, masterUid, skip, rows).Scan(&list)
	ctx.Resp.Json(dto.UserHomeVideoListRes{
		Endpoint: ctx.Config.Get("videobiz.qiniuObs.videoBucketUrl").ToString(),
		List:     list,
	})
}

// 获取用户点赞的视频
func SupportVideos(ctx *httpserver.Context) {
	db := orm.GetDB()
	masterUid := ctx.Req.GetInt("userId")
	page := ctx.Req.GetInt("page")
	rows := 12
	skip := (page - 1) * rows
	sql := "SELECT t.video_id, v.title, v.cover, v.support_num FROM video_support t " +
		"LEFT JOIN videos v ON t.video_id = v.id WHERE t.user_id = ? LIMIT ?, ?"
	list := []dto.UserHomeVideoItem{}
	db.Raw(sql, masterUid, skip, rows).Scan(&list)
	ctx.Resp.Json(dto.UserHomeVideoListRes{
		Endpoint: ctx.Config.Get("videobiz.qiniuObs.videoBucketUrl").ToString(),
		List:     list,
	})
}

// 获取用户收藏的视频
func CollectVideos(ctx *httpserver.Context) {
	db := orm.GetDB()
	masterUid := ctx.Req.GetInt("userId")
	page := ctx.Req.GetInt("page")
	rows := 12
	skip := (page - 1) * rows
	sql := "SELECT t.video_id, v.title, v.cover, v.support_num FROM video_collect t " +
		"LEFT JOIN videos v ON t.video_id = v.id WHERE t.user_id = ? LIMIT ?, ?"
	list := []dto.UserHomeVideoItem{}
	db.Raw(sql, masterUid, skip, rows).Scan(&list)
	ctx.Resp.Json(dto.UserHomeVideoListRes{
		Endpoint: ctx.Config.Get("videobiz.qiniuObs.videoBucketUrl").ToString(),
		List:     list,
	})
}

// 关注/取关
func Follow(ctx *httpserver.Context) {
	db := orm.GetDB()
	masterId := ctx.Req.GetInt("userId")
	myUid := ctx.GetVal("userId").ToString()
	id := 0
	var msg string
	var sql = "SELECT id FROM follow WHERE master = ? AND fans = ?"
	db.Raw(sql, masterId, myUid).Scan(&id)
	if id > 0 {
		sql = "DELETE FROM follow WHERE master = ? AND fans = " + myUid
		db.Exec(sql, masterId)
		// 维护 user 表计数
		db.Exec("UPDATE user SET follow_num = follow_num-1 WHERE id = ?", myUid)
		db.Exec("UPDATE user SET fans_num = fans_num-1 WHERE id = ?", masterId)
		msg = "unfollow success"
	} else {
		sql = "INSERT INTO follow SET master = ?, fans = " + myUid + ", create_time = ?"
		db.Exec(sql, masterId, timekit.NowDatetimeStr())
		// 维护 user 表计数
		db.Exec("UPDATE user SET follow_num = follow_num+1 WHERE id = ?", myUid)
		db.Exec("UPDATE user SET fans_num = fans_num+1 WHERE id = ?", masterId)
		msg = "follow success"
	}
	ctx.Resp.Json(dto.BaseRes{
		ApiMessage: msg,
	})
}

// 我关注的人
func FollowList(ctx *httpserver.Context) {
	db := orm.GetDB()
	myUid := ctx.GetVal("userId").ToString()
	page := ctx.Req.GetInt("page", 1)
	rows := 20
	skip := (page - 1) * rows
	sql := "SELECT u.nickname, u.avatar, u.id, t.create_time " +
		"FROM follow t LEFT JOIN user u ON t.master = u.id " +
		"WHERE fans = " + myUid + " ORDER BY id DESC LIMIT ?, ?"
	list := []dto.MyUserListItem{}
	db.Raw(sql, skip, rows).Scan(&list)
	ctx.Resp.Json(dto.MyUserListRes{
		List: list,
	})
}

// 我的粉丝
func FansList(ctx *httpserver.Context) {
	db := orm.GetDB()
	myUid := ctx.GetVal("userId").ToString()
	page := ctx.Req.GetInt("page", 1)
	rows := 20
	skip := (page - 1) * rows
	sql := "SELECT u.nickname, u.avatar, u.id, t.create_time " +
		"FROM follow t LEFT JOIN user u ON t.fans = u.id " +
		"WHERE master = " + myUid + " ORDER BY id DESC LIMIT ?, ?"
	list := []dto.MyUserListItem{}
	db.Raw(sql, skip, rows).Scan(&list)
	ctx.Resp.Json(dto.MyUserListRes{
		List: list,
	})
}

// 最近获赞
func SupportList(ctx *httpserver.Context) {
	db := orm.GetDB()
	myUid := ctx.GetVal("userId").ToString()
	page := ctx.Req.GetInt("page", 1)
	rows := 20
	skip := (page - 1) * rows
	sql := "SELECT distinct(u.id), u.nickname, u.avatar, u.id, t.create_time " +
		"FROM video_support t LEFT JOIN user u ON t.user_id = u.id " +
		"WHERE video_master_id = " + myUid + " ORDER BY id DESC LIMIT ?, ?"
	list := []dto.MyUserListItem{}
	db.Raw(sql, skip, rows).Scan(&list)
	ctx.Resp.Json(dto.MyUserListRes{
		List: list,
	})
}
