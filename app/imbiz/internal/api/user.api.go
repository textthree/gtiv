package api

import (
	"github.com/text3cn/goodle/kit/arrkit"
	"github.com/text3cn/goodle/kit/cryptokit"
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/kit/timekit"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"gtiv/app/imbiz/internal/caches"
	"gtiv/app/imbiz/internal/dto"
	"gtiv/app/imbiz/internal/entity"
	"gtiv/app/imbiz/internal/services"
	"gtiv/kit"
	"gtiv/kit/dbkit"
	"gtiv/types/errcode"
)

// @Summary 注册用户
// @Tags user
// @Router /user/register [post]
// @Param request body dto.UserRegisterReq true " "
// @Success 200 {object} dto.UserRegisterRes
func Register(ctx *httpserver.Context) {
	req := dto.UserRegisterReq{}
	ctx.Req.JsonScan(&req)
	if req.VerifyCode != 33333 {
		ctx.Resp.Json(dto.UserRegisterRes{
			BaseRes: dto.BaseRes{
				ApiCode:    errcode.Fail,
				ApiMessage: "Verify code error",
			},
		})
		return
	}

	data := dto.UserRegisterRes{}
	data.LastLoginTime = timekit.NowTimestamp()
	username := strkit.Tostring(req.TelCode) + strkit.Tostring(req.Tel)
	user := entity.User{
		Username:      username,
		TelCode:       req.TelCode,
		Tel:           strkit.Tostring(req.Tel),
		Password:      cryptokit.Md5(req.Password + "gtiv"),
		Nickname:      "+" + username,
		LastLoginTime: data.LastLoginTime,
	}
	db := orm.GetDB()
	res := db.Create(&user)
	if res.Error != nil {
		ctx.Resp.Json(dto.UserRegisterRes{
			BaseRes: dto.BaseRes{
				ApiCode:    errcode.Fail,
				ApiMessage: "The number is registered",
			},
		})
		return
	}
	data.Uid = strkit.Tostring(user.Id)
	data.Token = services.User(ctx).GenerateToken(data.Uid)
	ctx.Resp.Json(data)
}

// @Summary 登录
// @Tags user
// @Router /user/login [post]
// @Param request body dto.UserLoginReq true " "
// @Success 200 {object} dto.UserLoginRes
func UserLogin(ctx *httpserver.Context) {
	req := dto.UserLoginReq{}
	ctx.Req.JsonScan(&req)
	db := orm.GetDB()
	sql := "SELECT id as UserId, role as UserRole, nickname, username, avatar, gender, version, contacts_version, last_login_time, " +
		"fans_num, collect_num, support_num, create_video_num, follow_num " +
		"FROM user WHERE username = ? and password = ?"
	var userinfo dto.Userinfo
	db.Raw(sql, req.Username, cryptokit.Md5(req.Password+"gtiv")).Scan(&userinfo)
	if userinfo.UserId == 0 {
		ctx.Resp.Json(dto.UserRegisterRes{
			BaseRes: dto.BaseRes{
				ApiCode:    errcode.LoginAccountPasswordFail,
				ApiMessage: "Login faild",
			},
		})
		return
	}
	// 登录成功
	lastLoginTime := timekit.NowTimestamp()
	sql = "UPDATE user SET last_login_time = ? WHERE id = ?"
	db.Raw(sql, lastLoginTime, userinfo.UserId)
	// 重建缓存
	caches.RebuildUserinfoCache(ctx, strkit.Tostring(userinfo.UserId))
	userinfo.Token = services.User(ctx).GenerateToken(strkit.Tostring(userinfo.UserId))
	ctx.Resp.Json(dto.UserLoginRes{
		Userinfo: userinfo,
	})
}

// @Summary 获取用户信息
// @Tags user
// @Router /user/info [get]
// @Param request body dto.UserinfoReq true " "
// @Success 200 {object} dto.UserinfoRes
func UserInfo(ctx *httpserver.Context) {
	req := dto.UserinfoReq{}
	ctx.Req.JsonScan(&req)
	// FIXME: req.UserId 使用的是 Mysql 主键，容易被遍历，
	// 最好对其进行加密传输，即后端凡是返回 userId 的接口都用加密后的密文进行传输
	var userId string
	if req.UserId != "" {
		userId = req.UserId
	} else {
		userId = ctx.GetVal("uid").ToString()
	}
	userinfo := caches.GetUserinfo(ctx, userId)
	ctx.Resp.Json(userinfo)
}

// @Summary 修改用户信息
// @Tags user
// @Router /user/info [get]
// @Param request body dto.UpdateUserinfoReq true " "
// @Success 200 {object} dto.UpdateUserinfoRes
func UpdateUserInfo(ctx *httpserver.Context) {
	req := dto.UpdateUserinfoReq{}
	ctx.Req.JsonScan(&req)
	allowFields := []string{"avatar", "nickname"}
	if !arrkit.InArray(req.Field, allowFields) {
		kit.FailResponse(ctx, "Illegal field")
		return
	}
	services.User(ctx).EditUserInfo(req.Field, req.Value)
	kit.SuccessResponse(ctx)
}

// @Summary 查找用户
// @Tags user
// @Router /user/search [post]
// @Param request body dto.SearchUserReq true " "
// @Success 200 {object} dto.SearchUserRes
func SearchUser(ctx *httpserver.Context) {
	req := dto.SearchUserReq{}
	ctx.Req.JsonScan(&req)
	res, err := services.User(ctx).SearchUser(req)
	if err != nil {
		kit.FailResponse(ctx)
	}
	ctx.Resp.Json(res)
}

// @Summary 我关注的用户列表
// @Tags user
// @Router /user/subscribe_list [get]
// @Param request body dto.SubscribeUserListReq true " "
// @Success 200 {object} dto.SubscribeUserListRes
func SubscribeUserList(ctx *httpserver.Context) {
	req := dto.SubscribeUserListReq{}
	ctx.Req.JsonScan(&req)
	limit := dbkit.SqlLimitStatement(req.Page, req.Rows)
	list := services.User(ctx).SubscribeUserList(limit)
	ctx.Resp.Json(dto.SubscribeUserListRes{
		List: list,
	})
}
