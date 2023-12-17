package internal

import (
	"github.com/text3cn/goodle/providers/httpserver"
	goodlemiddleware "github.com/text3cn/goodle/providers/httpserver/middleware"
	"gtiv/app/imbiz/internal/api"
	"gtiv/app/imbiz/internal/middleware"
	"os"
	"path/filepath"
)

func Router(core *httpserver.Engine) {
	core.UseCross()
	dir, _ := os.Getwd()
	dir += string(filepath.Separator) + "i18n"
	core.UseMiddleware(middleware.Auth(), goodlemiddleware.I18n(dir))

	core.Get("/get-comet", api.GetComet)

	core.Get("/app/country", api.CountryList)
	core.Get("/app/version", api.Version)
	core.Get("/obstore/client_token", api.ClientToken)
	core.Post("/user/register", api.Register)
	core.Post("/user/login", api.UserLogin)
	core.Get("/user/info", api.UserInfo)
	core.Post("/user/info", api.UpdateUserInfo)
	core.Post("/user/search", api.SearchUser)
	core.Get("/user/subscribe_list", api.SubscribeUserList)

	core.Post("/contacts/add", api.AddContacts)
	core.Delete("/contacts/delete", api.DeleteContacts)
	core.Get("/contacts/list", api.ContactsList)

	// chat
	core.Post("/imlogic/push-mid", api.PushMid)
	core.Post("/imlogic/push-room", api.PushRoom)
	core.Post("/imlogic/check-online", api.CheckOnline)
	core.Get("/chat/addme-list", api.AddMeList)
	core.Get("/chat/sync-private-message", api.ChatRecord)
	core.Get("/chat/room-msg", api.GetRoomMessage)

	// room
	core.Post("/room/dissolve", api.DissolveRoom)
	core.Post("/room/modify_notice", api.RoomModifyNotice)
	core.Post("/room/modify_name", api.RoomModifyName)
	core.Post("/room/modify_avatar", api.RoomModifyAvatar)
	core.Post("/room/invite", api.InviteMember)
	core.Get("/room/list", api.RoomList)
	core.Get("/room/memberids", api.RoomMemberIds)
	core.Get("/room/get_notice", api.RoomGetNotice)
	core.Post("/room/banned_to_post", api.RoomBannedToPost)
	core.Post("/room/banned_list", api.RoomBannedList)
	core.Post("/room/relieve_banned_to_post", api.RoomRelieveBannedToPost)
	core.Delete("/room/remove_member", api.RemoveMember)
	core.Post("/room/admin_list", api.RoomAdminList)
	core.Get("/room/info", api.RoomInfo)
	core.Get("/room/member_info", api.RoomMemberInfo)
	core.Get("/room/member_list", api.RoomMemberList)
	core.Post("/room/set_admin", api.RoomSetAdmin)
	core.Post("/room/quit", api.QuitRoomReq)
	core.Get("/room/faq", api.RoomFaq)
}
