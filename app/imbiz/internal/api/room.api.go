package api

import (
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/providers/goodlog"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/redis"
	"gtiv/app/imbiz/internal/caches"
	"gtiv/app/imbiz/internal/caches/rediskey"
	"gtiv/app/imbiz/internal/constants"
	"gtiv/app/imbiz/internal/dto"
	"gtiv/app/imbiz/internal/services"
	"gtiv/kit"
	"gtiv/kit/dbkit"
	"time"
)

// @Summary 解散群
// @Tags imRoom
// @Router /room/dissolve [post]
// @Param request body dto.DissolveRoomReq true " "
// @Success 200 {object} dto.DissolveRoomRes
func DissolveRoom(ctx *httpserver.Context) {
	req := dto.DissolveRoomReq{}
	ctx.Req.JsonScan(&req)
	// 只有群主能解散属于自己的群
	errcode, errmsg := services.Room(ctx).CheckRole(req.RoomId, constants.RoomUserRole.RoomAdminOwner)
	if errcode > 0 {
		kit.FailResponse(ctx, errmsg, errcode)
		return
	}
	err := services.Room(ctx).DissolveRoom(req.RoomId)
	if err != nil {
		kit.FailResponse(ctx, err.Error())
		return
	}
	// 清理房间信息 redis
	conn := redis.Instance().Conn()
	roomInfoKey, _ := rediskey.RoomInfo(req.RoomId)
	conn.Do(ctx, "DEL", roomInfoKey)
	// 永久禁言全体成员
	services.Room(ctx).RoomBannedToPost(req.RoomId, "0", 3600*24*365)
	// 调用 goim 推送解散消息
	//url := config.BizCfg.Goim.Logic + "/goim/push/room?operation=" + req.RoomId +
	//	"&type=live&room=" + req.RoomId
	//url := ""
	//var message, toUser string
	//i18n := ctx.Value("i18n").(*gi18n.Manager)
	//
	//message = i18n.Translate(ctx, `dissolveRoom`)
	//toUser = "0"
	//msg := dto.ImMessage{
	//	RoomId:  req.RoomId,
	//	Content: message,
	//	MsgId:   uuid.New().String(),
	//	MsgType:    constants.MessageTypeBanToPost,
	//	ToUsers: toUser,
	//}
	//data := stringkit.JsonEncode(msg)
	//_, requstGoim := httpkit.HttpPostJson(url, data)
	//if requstGoim != nil {
	//	kit.Response(ctx, dto.BaseRes{
	//		ApiCode:    ErrorCode.Fail,
	//		ApiMessage: "goim error",
	//	})
	//	logger.Error(requstGoim.Error())
	//	return
	//}
	kit.SuccessResponse(ctx)
	return
}

// @Summary 修改群公告
// @Tags imRoom
// @Router /room/modify_notice [post]
// @Param request body dto.RoomModifyNoticeReq true " "
// @Success 200 {object} dto.RoomModifyNoticeRes
func RoomModifyNotice(ctx *httpserver.Context) {
	req := dto.RoomModifyNoticeReq{}
	ctx.Req.JsonScan(&req)
	// 权限判断
	services.Room(ctx).CheckRole(req.RoomId, constants.RoomUserRole.RoomAdmin)
	// 限制 1 - 200 字
	if len(req.Content) > 200 {
		kit.FailResponse(ctx, "The content limit is no more than 200 words ")
		return
	}
	services.Room(ctx).RoomModifyNotice(req.RoomId, req.Content)
	kit.SuccessResponse(ctx)
}

// @Summary 修改群名称
// @Tags imRoom
// @Router /room/modify_name [post]
// @Param request body dto.RoomModifyNameReq true " "
// @Success 200 {object} dto.RoomModifyNameRes
func RoomModifyName(ctx *httpserver.Context) {
	req := dto.RoomModifyNameReq{}
	ctx.Req.JsonScan(&req)
	services.Room(ctx).CheckRole(req.RoomId, constants.RoomUserRole.RoomAdmin)
	// 名称限制 20 字
	if req.Name == "" {
		kit.FailResponse(ctx, "The name cannot be empty")
		return
	}
	if len(req.Name) > 20 {
		kit.FailResponse(ctx, "The name limit is no more than 20 words")
		return
	}
	services.Room(ctx).RoomModifyName(req.RoomId, req.Name)
	kit.SuccessResponse(ctx)
}

// @Summary 修改群头像
// @Tags imRoom
// @Router /room/modify_avatar [post]
// @Param request body dto.RoomModifyAvatarReq true " "
// @Success 200 {object} dto.RoomModifyAvatarRes
func RoomModifyAvatar(ctx *httpserver.Context) {
	req := dto.RoomModifyAvatarReq{}
	ctx.Req.JsonScan(&req)
	services.Room(ctx).CheckRole(req.RoomId, constants.RoomUserRole.RoomAdmin)
	services.Room(ctx).RoomModifyAvatar(req.RoomId, req.Avatar)
	kit.SuccessResponse(ctx)
}

// @Summary 创建聊天室/邀请群成员
// @Tags imRoom
// @Router /room/invite [post]
// @Param request body dto.InviteMemberReq true " "
// @Success 200 {object} dto.InviteMemberRes
func InviteMember(ctx *httpserver.Context) {
	req := dto.InviteMemberReq{}
	ctx.Req.JsonScan(&req)
	conn := redis.Instance().Conn()
	createdRoomId, err := services.Room(ctx).InviteMember(req.RoomId, req.UserIds)
	if err != nil {
		goodlog.Error(err.Error())
		kit.FailResponse(ctx, err.Error())
		return
	}
	var message string
	var pusRoomId = req.RoomId
	var msgType int
	if req.RoomId == "" || req.RoomId == "0" {
		// 发送一条群消息
		message = ctx.I18n(`createNewRoom`)
		pusRoomId = createdRoomId
		msgType = constants.MessageTypeText
	} else {
		message = req.Message
		msgType = constants.InviteMember
	}
	services.Chat(ctx).PushRoom(&dto.PushRoomReq{
		Type:    msgType,
		Message: message,
		RoomId:  pusRoomId,
	})
	time.Sleep(time.Second * 1)
	// 通知客户端重新建立连接
	users := strkit.Explode(",", req.UserIds)
	ownerId := ctx.GetVal("uid").ToString()
	users = append(users, ownerId)
	for _, v := range users {
		if v != "" {
			services.Chat(ctx).PushMid(&dto.PushMidReq{
				Type:    constants.MessageNewRoom,
				ToUsers: v,
				Message: pusRoomId,
			})
		}
	}
	ctx.Resp.Json(dto.InviteMemberRes{
		RoomId:   pusRoomId,
		Message:  message,
		RoomName: ctx.I18n("whoseRoom"),
	})
	if req.RoomId != "" {
		// 清理房间信息 redis
		roomInfoKey, _ := rediskey.RoomInfo(req.RoomId)
		conn.Do(ctx, "DEL", roomInfoKey)
	}
}

// @Summary 我的群列表
// @Tags imRoom
// @Router /room/list [get]
// @Param request body dto.RoomListReq true " "
// @Success 200 {object} dto.RoomListRes
func RoomList(ctx *httpserver.Context) {
	req := dto.RoomListReq{}
	ctx.Req.JsonScan(&req)
	list := services.Room(ctx).RoomList()
	ctx.Resp.Json(dto.RoomListRes{List: list})
}

// @Summary 拉取所有已在群中的用户
// @Tags imRoom
// @Router /room/memberids [get]
// @Param request body dto.RoomMemberIdsReq true " "
// @Success 200 {object} dto.RoomMemberIdsRes
func RoomMemberIds(ctx *httpserver.Context) {
	req := dto.RoomMemberIdsReq{}
	ctx.Req.JsonScan(&req)
	list, _ := services.Room(ctx).RoomMemberIds(req.RoomId)
	ctx.Resp.Json(dto.RoomMemberIdsRes{
		List: list,
	})
}

// @Summary 查看群公告
// @Tags imRoom
// @Router /room/get_notice [get]
// @Param request body dto.RoomGetNoticeReq true " "
// @Success 200 {object} dto.RoomGetNoticeRes
func RoomGetNotice(ctx *httpserver.Context) {
	req := dto.RoomGetNoticeReq{}
	ctx.Req.JsonScan(&req)
	content, _ := services.Room(ctx).RoomGetNotice(req.RoomId)
	ctx.Resp.Json(dto.RoomGetNoticeRes{
		Content: content,
	})
}

// @Summary 禁言
// @Tags imRoom
// @Router /room/banned_to_post [post]
// @Param request body dto.RoomBannedToPostReq true " "
// @Success 200 {object} dto.RoomBannedToPostRes
func RoomBannedToPost(ctx *httpserver.Context) {
	req := dto.RoomBannedToPostReq{}
	ctx.Req.JsonScan(&req)
	second := 0
	timeText := ""
	// 禁言类型: 1、10分钟 2、1小时 3、12小时 4、1天 5、永久禁言
	switch req.Type {
	case 1:
		second = 600
		timeText = ctx.I18n("bannedTimeType1")
	case 2:
		second = 3600
		timeText = ctx.I18n("bannedTimeType2")
	case 3:
		second = 3600 * 12
		timeText = ctx.I18n("bannedTimeType3")
	case 4:
		second = 3600 * 24
		timeText = ctx.I18n("bannedTimeType4")
	case 5:
		second = 3600 * 24 * 365 // 永久禁言
		timeText = ctx.I18n("bannedTimeType5")
	}
	services.Room(ctx).RoomBannedToPost(req.RoomId, req.UserId, second)
	// 推送禁言消息
	var message, toUser string
	if req.UserId == "0" {
		message = ctx.I18n("bannedAll") // 全体禁言
		toUser = "0"
	} else {
		userinfo := caches.GetUserinfo(ctx, req.UserId)
		message = ctx.I18n(`user`) + userinfo.Nickname + timeText
		toUser = req.UserId
	}
	params := dto.PushRoomReq{}
	params.RoomId = req.RoomId
	params.Message = message
	params.Type = constants.MessageTypeBanToPost
	params.ToUsers = toUser
	_, _, err := services.Chat(ctx).PushRoom(&params)
	if err != nil {
		kit.FailResponse(ctx, err.Error())
		return
	}
	kit.SuccessResponse(ctx)
}

// @Summary 禁言列表
// @Tags imRoom
// @Router /room/banned_list [post]
// @Param request body dto.RoomBannedListReq true " "
// @Success 200 {object} dto.RoomBannedListRes
func RoomBannedList(ctx *httpserver.Context) {
	req := dto.RoomBannedListReq{}
	ctx.Req.JsonScan(&req)
	ret, _ := services.Room(ctx).RoomBannedList(req.RoomId)
	ctx.Resp.Json(dto.RoomBannedListRes{
		List: ret,
	})
}

// @Summary 解除禁言
// @Tags imRoom
// @Router /room/relieve_banned_to_post [post]
// @Param request body dto.RoomRelieveBannedToPostReq true " "
// @Success 200 {object} dto.RoomRelieveBannedToPostRes
func RoomRelieveBannedToPost(ctx *httpserver.Context) {
	req := dto.RoomRelieveBannedToPostReq{}
	ctx.Req.JsonScan(&req)
	services.Room(ctx).RoomRelieveBannedToPost(req.RoomId, req.UserId)
	var message, toUser string
	if req.UserId == "0" {
		message = ctx.I18n("relieveBannedAll") // 解除全体禁言
		toUser = "0"
	} else {
		userinfo := caches.GetUserinfo(ctx, req.UserId)
		message = ctx.I18n("user") + userinfo.Nickname + ctx.I18n("relieveBannedUser")
		toUser = req.UserId
	}
	// 推送消息
	params := dto.PushRoomReq{}
	params.RoomId = req.RoomId
	params.Message = message
	params.Type = constants.MessageTypeRelieveBan
	params.ToUsers = toUser
	_, _, err := services.Chat(ctx).PushRoom(&params)
	if err != nil {
		kit.FailResponse(ctx, err)
		return
	}
	kit.SuccessResponse(ctx)
}

// @Summary 移除群成员
// @Tags imRoom
// @Router /room/remove_member [delete]
// @Param request body dto.RemoveMemberReq true " "
// @Success 200 {object} dto.RemoveMemberRes
func RemoveMember(ctx *httpserver.Context) {
	req := dto.RemoveMemberReq{}
	ctx.Req.JsonScan(&req)
	// 至少是本群管理员才能操作
	services.Room(ctx).CheckRole(req.RoomId, constants.RoomUserRole.RoomAdmin)
	err := services.Room(ctx).RemoveMember(req.RoomId, req.UserId)
	if err != nil {
		kit.FailResponse(ctx, err.Error())
		return
	}
	// 清理房间信息 redis
	conn := redis.Instance().Conn()
	roomInfoKey, _ := rediskey.RoomInfo(req.RoomId)
	conn.Do(ctx, "DEL", roomInfoKey)
	kit.SuccessResponse(ctx)
}

// @Summary 群管理员列表
// @Tags imRoom
// @Router /room/admin_list [post]
// @Param request body dto.RoomAdminListReq true " "
// @Success 200 {object} dto.RoomAdminListRes
func RoomAdminList(ctx *httpserver.Context) {
	req := dto.RoomAdminListReq{}
	ctx.Req.JsonScan(&req)
	ret, err := services.Room(ctx).RoomAdminList(req.RoomId)
	if err != nil {
		kit.FailResponse(ctx)
		return
	}
	ctx.Resp.Json(dto.RoomAdminListRes{
		List: ret,
	})
}

// @Summary 获取群信息
// @Tags imRoom
// @Router /room/info [get]
// @Param request body dto.RoomInfoReq true " "
// @Success 200 {object} dto.RoomInfoRes
func RoomInfo(ctx *httpserver.Context) {
	req := dto.RoomInfoReq{}
	ctx.Req.JsonScan(&req)
	// 获取群信息
	ret, err := services.Room(ctx).RoomInfo(req.RoomId)
	if err != nil {
		kit.FailResponse(ctx)
		return
	}
	// 获取用户在这个群中的信息
	isMember, role, _ := services.Room(ctx).RoomUserinfo(req.RoomId)
	ret.IsMember = isMember
	ret.Role = role
	ctx.Resp.Json(ret)
}

// @Summary 常见问答列表
// @Tags imRoom
// @Router /room/faq [get]
// @Param request body dto.RoomFaqListReq true " "
// @Success 200 {object} dto.RoomFaqListRes
func RoomFaq(ctx *httpserver.Context) {
	list, _ := services.Room(ctx).RoomFaq()
	ctx.Resp.Json(dto.RoomFaqListRes{
		List: list,
	})
}

// @Summary 群成员信息获取
// @Tags imRoom
// @Router /room/member_info [get]
// @Param request body dto.RoomMemberInfoReq true " "
// @Success 200 {object} dto.RoomMemberInfoRes
func RoomMemberInfo(ctx *httpserver.Context) {
	req := dto.RoomMemberInfoReq{}
	ctx.Req.JsonScan(&req)
	ret := dto.RoomMemberInfoRes{}
	info := caches.GetUserinfo(ctx, req.UserId)
	ret.Avatar = info.Avatar
	ret.Nickname = info.Nickname
	_, role, createTime := services.Room(ctx).RoomUserinfo(req.RoomId, req.UserId)
	ret.CreateTime = createTime
	ret.Role = int(role)
	ctx.Resp.Json(ret)
}

// @Summary 群成员列表
// @Tags imRoom
// @Router /room/member_list [get]
// @Param request body dto.RoomMemberListReq true " "
// @Success 200 {object} dto.RoomMemberListRes
func RoomMemberList(ctx *httpserver.Context) {
	req := dto.RoomMemberListReq{}
	ctx.Req.JsonScan(&req)
	ret := dto.RoomMemberListRes{}
	limit := dbkit.SqlLimitStatement(req.Page, req.Rows)
	list := services.Room(ctx).RoomMemberList(req.RoomId, req.Nickname, limit)
	ret.List = list
	ctx.Resp.Json(ret)
}

// @Summary 设置管理员
// @Tags imRoom
// @Router /room/set_admin [post]
// @Param request body dto.RoomSetAdminReq true " "
// @Success 200 {object} dto.RoomSetAdminRes
func RoomSetAdmin(ctx *httpserver.Context) {
	req := dto.RoomSetAdminReq{}
	ctx.Req.JsonScan(&req)
	err := services.Room(ctx).RoomSetAdmin(req.RoomId, req.UserIds)
	if err != nil {
		kit.FailResponse(ctx)
		return
	}
	kit.SuccessResponse(ctx)
}

// @Summary 退出群聊
// @Tags imRoom
// @Router /room/quit [post]
// @Param request body dto.QuitRoomReq true " "
// @Success 200 {object} dto.QuitRoomRes
func QuitRoomReq(ctx *httpserver.Context) {
	req := dto.QuitRoomReq{}
	ctx.Req.JsonScan(&req)
	services.Room(ctx).QuitRoom(req.RoomId)
	// 清理房间信息 redis
	conn := redis.Instance().Conn()
	roomInfoKey, _ := rediskey.RoomInfo(req.RoomId)
	conn.Do(ctx, "DEL", roomInfoKey)
	kit.SuccessResponse(ctx)
}
