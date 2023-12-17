package api

import (
	"github.com/text3cn/goodle/kit/arrkit"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"gtiv/app/videobiz/internal/dto"
)

// @Summary 首页视频
// @Router /users/{id} [get]
// @Success 200 {object} dto.VideoHomeRes
func HomeVideo(ctx *httpserver.Context) {
	list := []dto.VideoHomeItem{}
	db := orm.GetDB()
	sql := "SELECT t.id, title, uri, cover, width, height, t.support_num, t.collect_num, t.share_num, " +
		"nickname, avatar, user_id FROM videos t " +
		"LEFT JOIN user ON t.user_id = user.id ORDER BY RAND() LIMIT 5"
	//"LEFT JOIN user ON t.user_id = user.id WHERE user.id = 228 LIMIT 5"
	db.Raw(sql).Scan(&list)
	ret := dto.VideoHomeRes{
		Endpoint: ctx.Config.Get("videobiz.qiniuObs.videoBucketUrl").ToString(),
		List:     list,
	}
	uid := ctx.GetVal("userId").ToInt()
	// 关注状态
	ids := []int{}
	sql = "SELECT master FROM follow WHERE fans = ?"
	db.Raw(sql, uid).Scan(&ids)
	for k, v := range ret.List {
		if arrkit.InArray(v.UserId, ids) || v.UserId == uid {
			ret.List[k].IsFollow = true
		}
	}
	ctx.Resp.Json(ret)
}
