package api

import (
	"github.com/text3cn/goodle/kit/arrkit"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"gtiv/app/videobiz/internal/dto"
)

func LiveList(ctx *httpserver.Context) {
	list := []dto.LiveListItem{}
	db := orm.GetDB()
	sql := "SELECT user_id, room_id, title, cover, state, nickname, avatar " +
		" FROM live_room t LEFT JOIN user ON t.user_id = user.id LIMIT 10"
	db.Raw(sql).Scan(&list)
	ret := dto.LiveListRes{
		Rtmp: ctx.Config.Get("videobiz.live_server.rtmp").ToString(),
		Hls:  ctx.Config.Get("videobiz.live_server.hls").ToString(),
		Flv:  ctx.Config.Get("videobiz.live_server.flv").ToString(),
		List: list,
	}
	uid := ctx.GetVal("userId").ToInt()
	// 关注状态
	ids := []int{}
	sql = "SELECT master FROM follow WHERE fans = ?"
	db.Raw(sql, uid).Scan(&ids)
	for k, v := range ret.List {
		if arrkit.InArray(v.UserId, ids) {
			ret.List[k].IsFollow = true
		}
	}
	ctx.Resp.Json(ret)
}
