package constants

// 群用户的群角色
type roomUserRole struct {
	Normal         int // 普通用户
	RoomAdmin      int // 群普通管理员
	RoomAdminOwner int // 群主
}

var RoomUserRole = roomUserRole{
	Normal:         10,
	RoomAdmin:      20,
	RoomAdminOwner: 30,
}

// 用户角色
type userRole struct {
	Normal      int // 普通用户
	SystemAdmin int // 系统管理员
}

var UserRole = userRole{
	Normal:      10,
	SystemAdmin: 20,
}
