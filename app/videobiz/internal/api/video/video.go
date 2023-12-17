package video

import (
	"github.com/text3cn/goodle/kit/timekit"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"gtiv/app/videobiz/internal/dto"
)

// 点赞/取消点赞
func Support(ctx *httpserver.Context) {
	db := orm.GetDB()
	videoId, _ := ctx.Req.FormInt("videoId")
	action, _ := ctx.Req.FormInt("action") // 1.点赞 0.取消
	myUid := ctx.GetVal("userId").ToString()
	var msg, sql string
	if action == 0 {
		sql = "DELETE FROM video_support WHERE user_id = ? AND video_id = ?"
		db.Exec(sql, myUid, videoId)
		// 维护 user 表计数
		db.Exec("UPDATE user SET support_num = support_num-1 WHERE id = ?", myUid)
		// 维护 video 表计数
		db.Exec("UPDATE videos SET support_num = support_num-1 WHERE id = ?", videoId)
		msg = "cancel support success"
	} else {
		sql = "INSERT INTO video_support SET user_id = ?, video_id = ?, create_time = ?"
		db.Exec(sql, myUid, videoId, timekit.NowDatetimeStr())
		// 维护 user 表计数
		db.Exec("UPDATE user SET support_num = support_num+1 WHERE id = ?", myUid)
		// 维护 video 表计数
		db.Exec("UPDATE videos SET support_num = support_num+1 WHERE id = ?", videoId)
		msg = "support success"
	}
	ctx.Resp.Json(dto.BaseRes{
		ApiMessage: msg,
	})
}

// 收藏/取消收藏
func Collect(ctx *httpserver.Context) {
	db := orm.GetDB()
	videoId, _ := ctx.Req.FormInt("videoId")
	action, _ := ctx.Req.FormInt("action") // 1.点赞 0.取消
	myUid := ctx.GetVal("userId").ToString()
	var msg, sql string
	if action == 0 {
		msg = "cancel collect success"
		sql = "DELETE FROM video_collect WHERE user_id = ? AND video_id = ?"
		db.Exec(sql, myUid, videoId)
		// 维护 user 表计数
		db.Exec("UPDATE user SET collect_num = collect_num-1 WHERE id = ?", myUid)
		// 维护 video 表计数
		db.Exec("UPDATE videos SET collect_num = collect_num-1 WHERE id = ?", videoId)
	} else {
		msg = "collect success"
		sql = "INSERT INTO video_collect SET user_id = ?, video_id = ?, create_time = ?"
		db.Exec(sql, myUid, videoId, timekit.NowDatetimeStr())
		// 维护 user 表计数
		db.Exec("UPDATE user SET collect_num = collect_num+1 WHERE id = ?", myUid)
		// 维护 video 表计数
		db.Exec("UPDATE videos SET collect_num = collect_num+1 WHERE id = ?", videoId)
	}
	ctx.Resp.Json(dto.BaseRes{
		ApiMessage: msg,
	})

}

// 增加分享次数
func IncrShareNum(ctx *httpserver.Context) {
	db := orm.GetDB()
	videoId, _ := ctx.Req.FormInt("videoId")
	var sql = "UPDATE videos SET share_num = share_num+1 WHERE id = ?"
	db.Exec(sql, videoId)
	ctx.Resp.Json(dto.BaseRes{})
}

// 获取单个视频信息
func Info(ctx *httpserver.Context) {
	db := orm.GetDB()
	ret := dto.VideoInfo{}
	myUid := ctx.GetVal("userId").ToInt()
	videoId := ctx.Req.GetInt("videoId")
	sql := "SELECT t.support_num, t.collect_num, t.share_num, t.cover VideoCover, t.width VideoWidth, t. height VideoHeight," +
		"u.id MasterUid, u.avatar MasterAvatar " +
		"FROM videos t LEFT JOIN user u ON t.user_id = u.id " +
		"WHERE t.id = ?"
	db.Raw(sql, videoId).Scan(&ret)
	ret.VideoCover = ctx.Config.Get("videobiz.qiniuObs.videoBucketUrl").ToString() + ret.VideoCover
	// 关注状态
	id := 0
	sql = "SELECT id FROM follow WHERE fans = ? AND master = ? LIMIT 1"
	db.Raw(sql, myUid, ret.MasterUid).Scan(&id)
	if id > 0 {
		ret.IsFollow = true
	}
	ctx.Resp.Json(ret)
}
