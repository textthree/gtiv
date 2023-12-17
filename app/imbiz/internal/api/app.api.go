package api

import (
	"github.com/text3cn/goodle/providers/etcd"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"gtiv/app/imbiz/boot"
	"gtiv/app/imbiz/internal/dto"
)

// @Summary 从 etcd 获取 Comet 列表
// @Tags app
// @Router /get-comet [get]
// @Success 200 {object} []string
func GetComet(ctx *httpserver.Context) {
	list := etcd.Instance().GetServices("im-comet")
	ctx.Resp.Json(list)
}

// @Summary 获取客户端最新版本号
// @Tags app
// @Router /app/version [get]
// @Param request body dto.VersionReq true " "
// @Success 200 {object} dto.VersionRes
func Version(ctx *httpserver.Context) {
	req := dto.VersionReq{}
	ctx.Req.JsonScan(&req)
	if req.Type == 1 {
		ctx.Resp.Json(dto.VersionRes{
			Version:      boot.ImbizCfg.GetString("clientVersion.androidVersion"),
			Url:          boot.ImbizCfg.GetString("clientVersion.apkUrl"),
			ForceUpgrade: boot.ImbizCfg.GetBool("clientVersion.forceUpgrade"),
		})
	}
}

// FIXME: 写死两种语言，这不是最终解决方案。
// @Summary 获取开放的地区列表
// @Tags app
// @Router /app/country [get]
// @Param request body dto.CountryReq true " "
// @Success 200 {object} dto.CountryRes
func CountryList(ctx *httpserver.Context) {
	language := ctx.GetVal("lng").ToString()
	sql := "SELECT name_en as country, tel_code FROM country WHERE open = 1 ORDER BY name_en"
	if language == "zh" {
		sql = "SELECT name as country, tel_code FROM country WHERE open = 1 ORDER BY CONVERT(name USING gbk)"
	}
	db := orm.GetDB()
	var list []dto.CountryItemRes
	db.Raw(sql).Scan(&list)
	ctx.Resp.Json(dto.CountryRes{
		List: list,
	})
}
