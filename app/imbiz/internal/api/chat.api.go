package api

import (
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/providers/httpserver"
	"gtiv/app/imbiz/im"
	"gtiv/app/imbiz/internal/dto"
	"gtiv/app/imbiz/internal/services"
	"gtiv/kit"
)

// @Summary 发送私信
// @Tags chat
// @Router /imlogic/push-mid [post]
// @Param request body dto.PushMidReq true " "
// @Success 200 {object} dto.PushMidRes
func PushMid(ctx *httpserver.Context) {
	req := dto.PushMidReq{}
	ctx.Req.JsonScan(&req)
	serverMsgId, time, err := services.Chat(ctx).PushMid(&req)
	if err != nil {
		kit.FailResponse(ctx, err.Error())
		return
	}
	ctx.Resp.Json(dto.PushMidRes{
		Time:        time,
		ServerMsgId: serverMsgId,
	})
	return
}

// @Summary 发群聊消息
// @Tags chat
// @Router /imlogic/check-online [post]
// @Param request body dto.CheckOnlineReq true " "
// @Success 200 {object} dto.CheckOnlineRes
func PushRoom(ctx *httpserver.Context) {
	req := dto.PushRoomReq{}
	ctx.Req.JsonScan(&req)
	serverMsgId, time, err := services.Chat(ctx).PushRoom(&req)
	if err != nil {
		kit.FailResponse(ctx, err.Error())
		return
	}
	ctx.Resp.Json(dto.PushRoomRes{Time: time, ServerMsgId: serverMsgId})
}

// @Summary 判断用户是否在线
// @Tags chat
// @Router /imlogic/push-room [post]
// @Param request body dto.PushRoomReq true " "
// @Success 200 {object} dto.PushRoomRes
func CheckOnline(ctx *httpserver.Context) {
	req := dto.CheckOnlineReq{}
	ctx.Req.JsonScan(&req)
	uid := req.User
	if req.User == "" {
		uid = ctx.GetVal("uid").ToString()
	}
	mids := []int64{strkit.ParseInt64(uid)}
	isOnline := im.ImLogic.CheckOnline(ctx, mids)
	ctx.Resp.Json(dto.CheckOnlineRes{Online: isOnline})
}

// @Summary 请求加我为好友的人
// @Tags chat
// @Router /chat/addme-list [get]
// @Param request body dto.AddMeListReq true " "
// @Success 200 {object} dto.AddMeListRes
func AddMeList(ctx *httpserver.Context) {
	list, err := services.Chat(ctx).AddMeList()
	if err != nil {
		kit.FailResponse(ctx)
		return
	}
	ctx.Resp.Json(dto.AddMeListRes{
		List: list,
	})
}

// @Summary 获取未同步的私聊聊天记录
// @Tags chat
// @Router /chat/sync-private-message [get]
// @Param request body dto.ChatRecordReq true " "
// @Success 200 {object} dto.ChatRecordRes
func ChatRecord(ctx *httpserver.Context) {
	req := dto.ChatRecordReq{}
	ctx.Req.JsonScan(&req)
	if req.LastMessageTime == 0 {
		return
	}
	list, err := services.Chat(ctx).ChatRecord(req.LastMessageTime)
	if err != nil {
		kit.FailResponse(ctx)
		return
	}
	ctx.Resp.Json(dto.ChatRecordRes{
		List: list,
	})
}

// @Summary 同步单个群聊消息
// @Tags chat
// @Router /chat/room-msg [get]
// @Param request body dto.RoomMsgReq true " "
// @Success 200 {object} dto.RoomMsgRes
func GetRoomMessage(ctx *httpserver.Context) {
	req := dto.RoomMsgReq{}
	ctx.Req.JsonScan(&req)

	// 检查是否群成员
	isMember, _, _ := services.Room(ctx).RoomUserinfo(req.RoomId)
	if !isMember {
		kit.FailResponse(ctx, "you are not member")
		return
	}
	if req.LastMessageTime == 0 {
		// 为 0 不能同步，防止删除消息后重新卸载安装 App 又同步到数据
		// 如果客户端通过接受推送消息后得到最新的 LastMessageTime 也不靠谱，如果客户端你一直不在线就一直推送不到,
		// 解决方案是后端帮用户记录同步时间，客户端传递 0 时使用服务端时间去同步,
		// 目前没时间搞先用这个不靠谱的做法，为 0 时也会同步数据，让它重复同步先吧，后面当 bug 修
		kit.SuccessResponse(ctx)
		return
	}
	list, err := services.Chat(ctx).RoomMsg(req.RoomId, int64(req.LastMessageTime))
	if err != nil {
		kit.FailResponse(ctx)
		return
	}
	ctx.Resp.Json(dto.RoomMsgRes{
		List: list,
	})

}
