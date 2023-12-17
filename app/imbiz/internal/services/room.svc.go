package services

import (
	"encoding/json"
	"errors"
	"github.com/spf13/cast"
	"github.com/text3cn/goodle/kit/castkit"
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/kit/timekit"
	"github.com/text3cn/goodle/providers/goodlog"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"github.com/text3cn/goodle/providers/redis"
	"gorm.io/gorm"
	"gtiv/app/imbiz/internal/caches"
	"gtiv/app/imbiz/internal/caches/rediskey"
	"gtiv/app/imbiz/internal/constants"
	"gtiv/app/imbiz/internal/dto"
	"gtiv/app/imbiz/internal/entity"
	"gtiv/types/errcode"
	"time"
)

type room struct {
	db       *gorm.DB
	uid      *castkit.GoodleVal
	ctx      *httpserver.Context
	userinfo *dto.Userinfo
}

var roomInstance *room

func Room(ctx *httpserver.Context) *room {
	if roomInstance != nil {
		return roomInstance
	}
	uid := ctx.GetVal("uid")
	userinfo := caches.GetUserinfo(ctx, uid.ToString())
	return &room{
		db:       orm.GetDB(),
		uid:      uid,
		ctx:      ctx,
		userinfo: userinfo,
	}
}

// 获取群信息
func (this room) RoomInfo(roomId string) (ret dto.RoomInfoRes, err error) {
	// 全员禁言结束时间
	conn := redis.Instance().Conn()
	ret.BannedAll = caches.CheckRoomBannedExpire(roomId, "0")
	// 用户个人禁言结束时间
	ret.BannedUser = caches.CheckRoomBannedExpire(roomId, this.uid.ToString())
	// 群信息
	roomInfoKey, expire := rediskey.RoomInfo(roomId)
	cache, _ := conn.Do(this.ctx, "GET", roomInfoKey).Result()
	if cache != nil {
		data, err := strkit.Json_decode_map(cast.ToString(cache))
		if err != nil {
			goodlog.Error("群信息反序列化出错")
		}
		id, _ := data["room_id"].(string)
		avatar, _ := data["avatar"].(string)
		memberNum, _ := data["member_num"].(json.Number).Int64()
		roomName, _ := data["room_name"].(string)
		roomType, _ := data["room_type"].(json.Number).Int64()
		ret.Id = id
		ret.Avatar = avatar
		ret.MemberNum = memberNum
		ret.RoomName = roomName
		ret.RoomType = roomType
	} else {
		sql := "SELECT id, avatar, room_name, member_num " +
			"FROM room WHERE id = ? "
		type fields struct {
			Id        string
			Avatar    string
			RoomName  string
			MemberNum int64
		}
		var row fields
		orm.GetDB().Raw(sql, roomId).Scan(&row)
		ret.Id = row.Id
		ret.Avatar = row.Avatar
		ret.MemberNum = row.MemberNum
		ret.RoomName = row.RoomName
		// 更新缓存
		cacheData := map[string]interface{}{
			"room_id":    ret.Id,
			"avatar":     ret.Avatar,
			"member_num": ret.MemberNum,
			"room_name":  ret.RoomName,
		}
		conn.Set(this.ctx, roomInfoKey, cacheData, expire)
		conn.Do(this.ctx, "EXPIRE", roomInfoKey, expire)
	}
	return
}

// 获取用户在某个群中的信息
// args[0] 用户 id
// @return isMember 是否群成员； role 在群中的角色
func (this room) RoomUserinfo(roomId string, args ...string) (isMember bool, role int64, createTime int) {
	userId := this.uid.ToString()
	if len(args) > 0 {
		userId = args[0]
	}
	conn := redis.Instance().Conn()

	key := rediskey.RoomUserInfo(roomId, userId)
	cache := conn.Get(this.ctx, key).Val()
	if cache != "" {
		data, err := strkit.Json_decode_map(cache)
		if err != nil {
			goodlog.Error("RoomUserinfo 反序列化出错")
		}
		isMember, _ = data["isMember"].(bool)
		roleF, _ := data["role"].(float64)
		role = int64(roleF)
		tF, _ := data["createTime"].(float64)
		t := int64(tF)
		createTime = int(t)
		return
	}
	sql := "SELECT id, role, create_time FROM user_room WHERE user_id = " + userId + " AND room_id = ?"
	type fields struct {
		Id         int
		Role       int64
		CreateTime time.Time
	}
	var row fields
	this.db.Raw(sql, roomId).Scan(&row)
	if row.Id > 0 {
		isMember = true
		role = row.Role
	} else {
		isMember = false
		role = int64(constants.RoomUserRole.Normal)
	}
	createTime = int(timekit.Time2stamp(row.CreateTime))
	cacheData := map[string]interface{}{
		"isMember":   isMember,
		"createTime": createTime,
		"role":       role,
	}
	conn.Set(this.ctx, key, strkit.JsonEncode(cacheData), time.Second*600) // 缓存十分钟
	return
}

// 判断用户是否是否具有某项群管理权限
// roomId 群 id
// role 该项操作需要达到的角色级别
func (this room) CheckRole(roomId string, role int) (errorCode int, errMsg string) {
	// 首先判断是否系统管理员，系统管理员直接放行
	if this.userinfo.UserRole == constants.UserRole.SystemAdmin {
		return
	}
	// 判断在该群的权限
	sql := "SELECT id, role as UserRole FROM user_room WHERE user_id = " + this.uid.ToString() + " AND room_id = ?"
	var id, UserRole int
	this.db.Raw(sql, roomId).Row().Scan(&id, &UserRole)
	if id == 0 || UserRole < role {
		// 不是群成员
		errorCode = errcode.NoPermission
		errMsg = "no permission"
	}
	return
}

// 解散群
func (this room) DissolveRoom(roomId string) (err error) {
	sql := "DELETE FROM user_room WHERE room_id = ?"
	this.db.Exec(sql, roomId)
	sql = "DELETE FROM room WHERE id = ?"
	this.db.Exec(sql, roomId)
	return
}

// 禁言
// roomId 房间
// userId 被禁的用户
// second 禁言多少秒
func (this room) RoomBannedToPost(roomId, userId string, second int) error {
	conn := redis.Instance().Conn()
	key := rediskey.RoomBanned(roomId)
	res := conn.Do(this.ctx, "HSET", key, userId, timekit.NowTimestamp()+second)
	return res.Err()
}

// 修改群公告
func (this room) RoomModifyNotice(roomId, notice string) {
	sql := "UPDATE room SET notice = ? " + " WHERE id = " + roomId
	this.db.Exec(sql, notice)
	conn := redis.Instance().Conn()
	key := rediskey.RoomNotice(roomId)
	conn.Do(this.ctx, "SET", key, notice)
}

// 修改群名称
func (this room) RoomModifyName(roomId, name string) {
	sql := "UPDATE room SET room_name = ? " + " WHERE id = " + roomId
	this.db.Exec(sql, name)
	// 清理群信息缓存
	conn := redis.Instance().Conn()
	key, _ := rediskey.RoomInfo(roomId)
	conn.Do(this.ctx, "DEL", key)
}

// 修改群头像
func (this room) RoomModifyAvatar(roomId, name string) {
	sql := "UPDATE room SET avatar = ? " + " WHERE id = " + roomId
	this.db.Exec(sql, name)
	// 清理群信息缓存
	conn := redis.Instance().Conn()
	key, _ := rediskey.RoomInfo(roomId)
	conn.Do(this.ctx, "DEL", key)
}

// 创建群/邀请成员
// roomId 房间 id，为空代表创建新群
// userIds 邀请进入房间的用户id，多个逗号隔开
func (this room) InviteMember(roomId, userIds string) (retRoomId string, err error) {
	usersArr := strkit.Explode(",", userIds)
	retRoomId = roomId
	if roomId == "" || roomId == "0" {
		if len(usersArr) > 500 {
			err = errors.New(this.ctx.I18n("userRoomMemberLimit"))
			return
		}
		// 限制每个用户只能创建 100 个群
		sql := "SELECT count(*) count FROM user_room t " +
			"LEFT JOIN user on t.user_id = user.id " +
			"WHERE t.user_id = ? AND t.role = ?"
		var count int
		this.db.Raw(sql, this.uid, constants.RoomUserRole.RoomAdminOwner)
		if count >= 100 {
			message := this.ctx.I18n(`createRooomCount1`) + " 100 " +
				this.ctx.I18n(`createRooomCount2`)
			err = errors.New(message)
			return
		}
		// 创建群
		newRoom := entity.Room{
			RoomName:  this.ctx.I18n("whoseRoom"),
			MemberNum: len(usersArr) + 1, // 加自己算一个
		}
		result := this.db.Create(&newRoom)
		if result.Error != nil {
			goodlog.Error(result.Error)
		}
		roomId = strkit.Tostring(newRoom.Id)
		retRoomId = roomId
		// 把自己加入群
		usersArr = append(usersArr, this.uid.ToString())

	} else {
		sql := "SELECT member_num FROM room WHERE id = " + roomId
		var memberNum int
		this.db.Raw(sql).Row().Scan(&memberNum)
		if memberNum+len(usersArr) > 500 {
			err = errors.New(this.ctx.I18n("userRoomMemberLimit"))
			return
		}
		// 已存在的群邀请人，更新群人数
		// 已在群中的人在前端不可选择，等于前端去重了，这里没有做判断去重了
		// 虽然依赖前端做逻辑判断不安全，但是先用着。
		sql = "UPDATE room SET member_num = member_num + " + strkit.Tostring(len(usersArr))
		this.db.Exec(sql)
	}
	// 拉人入群
	for _, v := range usersArr {
		role := constants.RoomUserRole.Normal
		if v == this.uid.ToString() {
			role = constants.RoomUserRole.RoomAdminOwner
		}
		userRoom := entity.UserRoom{
			UserId: cast.ToInt(retRoomId),
			RoomId: cast.ToInt(v),
			Role:   cast.ToInt8(role),
		}
		this.db.Create(&userRoom)
	}
	return
}

// 我的群列表
func (this room) RoomList() (ret []dto.RoomList) {
	sql := `SELECT room_id, room.room_name, room.member_num, room.member_limit, room.avatar
		    FROM user_room t LEFT JOIN room ON t.room_id = room.id
	      	WHERE t.user_id = ?`
	this.db.Raw(sql, this.uid.ToString()).Scan(&ret)
	return
}

// 拉取所有在群中的用户 id，只针对用户创建的群，用户群最多 500 人
func (this room) RoomMemberIds(roomId string) (rows []int, err error) {
	sql := "SELECT user_id FROM user_room WHERE room_id = ? LIMIT 500"
	this.db.Raw(sql, roomId).Scan(&rows)
	return
}

func (this room) RoomGetNotice(roomId string) (ret string, err error) {
	conn := redis.Instance().Conn()
	key := rediskey.RoomNotice(roomId)
	res := conn.Do(this.ctx, "GET", key)
	if res.Val() != nil {
		ret, _ = res.Val().(string)
		return
	}
	sql := "SELECT notice FROM room WHERE id = ? "
	this.db.Raw(sql, roomId).Scan(&ret)
	conn.Set(this.ctx, key, ret, time.Second*3600*3)
	return
}

// faq 列表
func (this room) RoomFaq() (ret []dto.RoomFaqItem, err error) {
	sql := "SELECT title, content FROM faq"
	this.db.Raw(sql).Scan(&ret)
	return
}

// 退出群
func (this room) QuitRoom(roomId string) {
	_sql := "DELETE FROM user_room WHERE room_id = ? AND user_id = ?"
	this.db.Exec(_sql, roomId, this.uid.ToString())
	_sql = "UPDATE room SET member_num = member_num - 1 WHERE id = ?"
	this.db.Exec(_sql, roomId)
}

// 设置管理员
func (this room) RoomSetAdmin(roomId, userIds string) (err error) {
	idsArr := strkit.Explode(",", userIds)
	// 先清空原来的管理员数据及这些管理员的缓存
	conn := redis.Instance().Conn()
	sql := "SELECT user_id ids FROM user_room WHERE room_id = ? AND role = ?"
	var ids []string
	this.db.Raw(sql, roomId, constants.RoomUserRole.RoomAdmin).Scan(&ids)
	if len(ids) > 0 {
		for _, v := range ids {
			roomInfoKey := rediskey.RoomUserInfo(roomId, v)
			conn.Do(this.ctx, "DEL", roomInfoKey)
		}
	}
	sql = "UPDATE user_room SET role =? WHERE room_id = ? AND role = ?"
	this.db.Exec(sql, constants.RoomUserRole.Normal, roomId, constants.RoomUserRole.RoomAdmin)
	// 然后重新设置并清理新设置管理员的缓存
	for _, v := range idsArr {
		if v == "" {
			continue
		}
		sql = "UPDATE user_room SET role =? WHERE room_id = ? AND user_id = ?" +
			" AND role = ?"
		this.db.Exec(sql, constants.RoomUserRole.RoomAdmin, roomId, v, constants.RoomUserRole.Normal)
		if err != nil {
			goodlog.Error(err)
		}
	}
	for _, v := range idsArr {
		roomInfoKey := rediskey.RoomUserInfo(roomId, v)
		conn.Do(this.ctx, "DEL", roomInfoKey)
	}
	return
}

// 获取群成员列表
func (this room) RoomMemberList(roomId, nickname string, limit string) (list []dto.RoomMemberListItem) {
	like := ""
	if nickname != "" {
		like = " AND u.nickname LIKE '%" + nickname + "%'"
	}
	sql := "SELECT t.user_id, u.nickname, t.create_time, t.role, u.avatar FROM user_room t " +
		" LEFT JOIN user u ON t.user_id = u.id " +
		" WHERE t.room_id = " + roomId + like +
		" ORDER BY t.role DESC " + limit
	goodlog.Yellow(sql)
	this.db.Raw(sql).Scan(&list)
	return
}

// 管理员列表
func (this room) RoomAdminList(roomId string) (ret []dto.RoomAdminList, err error) {
	sql := "SELECT t.user_id, t.role, u.nickname, u.avatar FROM user_room t" +
		" LEFT JOIN user u ON t.user_id = u.id " +
		" WHERE t.room_id = ? AND t.role > ?" +
		" ORDER BY t.role DESC"
	tx := this.db.Raw(sql, roomId, constants.RoomUserRole.Normal).Scan(&ret)
	if tx.Error != nil {
		err = tx.Error
		return
	}
	return
}

// 移除群成员（控制器中有做是否该群管理员检查，这里直接做移除逻辑）
func (this room) RemoveMember(roomId, userId string) (err error) {
	// 任何人不能移除群主，管理员不能移除管理员，即只能移除 role 币自己低的用户
	sql := "SELECT role FROM user_room WHERE user_id = " + this.uid.ToString()
	var role int
	this.db.Raw(sql).Scan(&role)
	myRole := role
	sql = "SELECT role FROM user_room WHERE user_id = ?"
	this.db.Raw(sql, userId).Scan(&role)
	userRole := role
	if myRole <= userRole {
		msg := this.ctx.I18n("canNotRemoveMember")
		err = errors.New(msg)
		return
	}
	// 移除
	sql = "DELETE FROM user_room WHERE room_id = ? AND user_id = ?"
	this.db.Exec(sql, roomId, userId)
	sql = "UPDATE room SET member_num = member_num - 1 WHERE id = ?"
	this.db.Exec(sql, roomId)
	return
}

// 禁言列表
func (this room) RoomBannedList(roomId string) (ret []dto.RoomBannedList, err error) {
	conn := redis.Instance().Conn()
	key := rediskey.RoomBanned(roomId)
	list := conn.Do(this.ctx, "HGETALL", key).Val()
	res, _ := list.(map[interface{}]interface{})
	userIds := ""
	for k, v := range res {
		uid := cast.ToString(k)
		if uid == "0" {
			continue
		}
		expire, _ := v.(string)
		expireTime := strkit.ParseInt(expire)
		// 检查是否禁言到期
		caches.CheckRoomBannedExpire(roomId, uid)
		item := dto.RoomBannedList{
			UserId: uid,
			Expire: expireTime,
		}
		ret = append(ret, item)
		userIds += uid + ","
	}
	userIds = strkit.TrimComma(userIds, "right")
	if userIds == "" {
		return
	}
	// 查询用户信息
	sql := "SELECT id, nickname, avatar FROM user WHERE id IN(" + userIds + ")"
	type fields struct {
		id       int
		nickname string
		avatar   string
	}
	var users []fields
	this.db.Raw(sql).Scan(&users)
	count := len(users)
	for i := 0; i < count; i++ {
		it := users[i]
		ret[i].Nickname = it.nickname
		ret[i].Avatar = it.avatar
	}
	return
}

// 解除禁言
func (this room) RoomRelieveBannedToPost(roomId, userId string) error {
	conn := redis.Instance().Conn()
	key := rediskey.RoomBanned(roomId)
	res := conn.Do(this.ctx, "HDEL", key, userId)
	return res.Err()
}
