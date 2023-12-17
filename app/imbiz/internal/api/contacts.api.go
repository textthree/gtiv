package api

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/kit/timekit"
	"github.com/text3cn/goodle/providers/goodlog"
	"github.com/text3cn/goodle/providers/httpserver"
	"gtiv/app/imbiz/im"
	"gtiv/app/imbiz/internal/constants"
	"gtiv/app/imbiz/internal/dto"
	"gtiv/app/imbiz/internal/services"
	"gtiv/kit"
	"gtiv/kit/impkg/protocol"
)

// @Summary 删除联系人
// @Tags contacts
// @Router /contacts/delete [delete]
// @Param request body dto.DeleteContactsReq true " "
// @Success 200 {object} dto.DeleteContactsRes
func DeleteContacts(ctx *httpserver.Context) {
	req := dto.DeleteContactsReq{}
	ctx.Req.JsonScan(&req)
	services.Contacts(ctx).DeleteContacts(req.UserId)
	kit.SuccessResponse(ctx)
}

// @Summary 联系人列表
// @Tags contacts
// @Router /contacts/list [get]
// @Param request body dto.ContactsListReq true " "
// @Success 200 {object} dto.ContactsListRes
func ContactsList(ctx *httpserver.Context) {
	res, err := services.Contacts(ctx).ContactsList()
	if err != nil {
		kit.FailResponse(ctx, err.Error())
	}
	ctx.Resp.Json(res)
}

// @Summary 同意或拒绝加联系人
// @Tags contacts
// @Router /contacts/add [post]
// @Param request body dto.AddContactsReq true " "
// @Success 200 {object} dto.AddContactsRes
func AddContacts(ctx *httpserver.Context) {
	goodlog.Pink("fjdslkfs s")
	req := dto.AddContactsReq{}
	ctx.Req.JsonScan(&req)
	res, err := services.Contacts(ctx).AddContacts(&req)
	myUid := ctx.GetVal("uid").ToString()
	if req.Type == 1 && err == nil {
		// 通知对方我已经通过好友验证
		mids := []int64{strkit.ParseInt64(req.UserId)}
		msg := dto.PushMidMessageDto{
			ChatType:    1,
			MsgType:     constants.MessageNewChat,
			FromUser:    myUid,
			ToUsers:     req.UserId,
			Time:        timekit.Millisecond(),
			ServerMsgId: uuid.New().String(),
			Content:     ctx.I18n("acceptContacts"),
			BadgeNum:    1,
		}
		msgBodyJson, err := json.Marshal(msg)
		if err == nil {
			im.ImLogic.PushToMids(ctx, protocol.OpMidMsg, mids, msgBodyJson)
		} else {
			goodlog.Error("[通知对方验证通过时出错]", err.Error())
		}
	}
	if err != nil {
		kit.FailResponse(ctx, err.Error())
		return
	}
	ctx.Resp.Json(res)
}
