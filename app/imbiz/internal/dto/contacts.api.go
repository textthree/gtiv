package dto

// 同意/拒绝加好友
type AddContactsReq struct {
	UserId string // 请求加我为好友的用户 id
	Type   int8   `swaggertype:"kotlin.Int"` // 1.接受 0.拒绝
}
type AddContactsRes struct {
	BaseRes
}

// 删除联系人
type DeleteContactsReq struct {
	UserId string // 用户 id
}
type DeleteContactsRes struct {
	BaseRes
}

// 联系人列表
type ContactsListReq struct {
}
type ContactsItem struct {
	UserId   string
	Username string
	Nickname string
	Avatar   string
	Gender   int
	Deleted  int
}
type ContactsListRes struct {
	BaseRes
	List []ContactsItem
}
