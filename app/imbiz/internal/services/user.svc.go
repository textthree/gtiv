package services

import (
	"github.com/text3cn/goodle/kit/castkit"
	"github.com/text3cn/goodle/kit/cryptokit"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"gorm.io/gorm"
	"gtiv/app/imbiz/boot"
	"gtiv/app/imbiz/internal/caches"
	"gtiv/app/imbiz/internal/dto"
)

type user struct {
	db  *gorm.DB
	uid *castkit.GoodleVal
	ctx *httpserver.Context
}

var userInstance *user

func User(ctx *httpserver.Context) *user {
	if userInstance != nil {
		return userInstance
	}
	return &user{
		db:  orm.GetDB(),
		uid: ctx.GetVal("uid"),
		ctx: ctx,
	}

}

// 生成 token
func (self *user) GenerateToken(userId string) string {
	return cryptokit.DynamicEncrypt(boot.TokenKey, userId)
}

// 修改用户信息
// TODO: 用户资料修改后发送一条 IM 消息通知所有好友
func (this user) EditUserInfo(field, value string) (err error) {
	sql := "UPDATE user SET " + field + " = ? WHERE id = ?"
	uid := this.uid.ToString()
	this.db.Exec(sql, value, uid)
	caches.RebuildUserinfoCache(this.ctx, uid)
	return
}

// 查找用户
func (this user) SearchUser(req dto.SearchUserReq) (ret dto.SearchUserRes, err error) {
	sql := "SELECT id as UserId, username, nickname, avatar" +
		" FROM user WHERE username = ?"
	//db.SetDebug(true)
	this.db.Raw(sql, req.Username).Scan(&ret)
	return
}

// 我关注的用户列表
func (this user) SubscribeUserList(limit string) (ret []dto.SubscribeUserList) {
	sql := "SELECT u.id UserId, u.nickname, u.avatar FROM follow " +
		" LEFT JOIN user u ON follow.master = u.id " +
		" WHERE fans = " + this.uid.ToString() + limit
	var rows []dto.SubscribeUserList
	this.db.Raw(sql).Scan(&rows)
	for _, item := range rows {
		if item.UserId == this.uid.ToString() {
			// 排除自己
			continue
		}
		ret = append(ret, item)
	}
	return
}
