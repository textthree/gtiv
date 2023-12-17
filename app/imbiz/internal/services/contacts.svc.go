package services

import (
	"errors"
	"github.com/text3cn/goodle/kit/castkit"
	"github.com/text3cn/goodle/providers/httpserver"
	"github.com/text3cn/goodle/providers/orm"
	"github.com/text3cn/goodle/providers/redis"
	"gorm.io/gorm"
	"gtiv/app/imbiz/internal/caches"
	"gtiv/app/imbiz/internal/caches/rediskey"
	"gtiv/app/imbiz/internal/dto"
)

type contacts struct {
	db  *gorm.DB
	uid *castkit.GoodleVal
	ctx *httpserver.Context
}

var contactsInstance *contacts

func Contacts(ctx *httpserver.Context) *contacts {
	if contactsInstance != nil {
		return contactsInstance
	}
	return &contacts{
		db:  orm.GetDB(),
		uid: ctx.GetVal("uid"),
		ctx: ctx,
	}

}

// 同意或拒绝添加联系人
func (this contacts) AddContacts(req *dto.AddContactsReq) (ret dto.AddContactsRes, err error) {
	conn := redis.Instance().Conn()
	myUid := this.uid.ToString()
	if req.Type == 1 {
		// 将我加入到他的通讯录中
		err = this.addContacts(myUid, req.UserId)
		// 将他加入到我的通讯录中
		err = this.addContacts(req.UserId, myUid)
	}
	// 删除验证列表中的项
	key, _ := rediskey.ApplyContactMeList(this.uid.ToString())
	conn.Do(this.ctx, "HDEL", key, req.UserId)
	return
}

func (this contacts) addContacts(myUserId string, contactsId string) (err error) {
	// 加入到我的通讯录中
	type result struct {
		Id     int
		Delete int
	}
	var res result
	sql := "SELECT id, deleted FROM contacts WHERE user_id = ? AND contacts_user_id = ?"
	this.db.Raw(sql, myUserId, contactsId).Scan(&res)
	if res.Id > 0 {
		sql = "INSERT INTO contacts SET user_id = ?, contacts_user_id = ?"
		this.db.Exec(sql, myUserId, contactsId)
	} else {
		if res.Delete == 1 {
			sql = "UPDATE contacts SET deleted = 0 WHERE user_id = ? AND contacts_user_id = ?"
			this.db.Exec(sql, myUserId, contactsId)
		} else {
			err = errors.New("already in contacts")
			return
		}
	}
	// 维护我的联系人版本号
	sql = "UPDATE user SET contacts_version = contacts_version + 1 WHERE id = " + myUserId
	this.db.Raw(sql)
	caches.RebuildUserinfoCache(this.ctx, myUserId)

	// 为减少 redis 压力，是否好友关系的判断暂且使用客户端判断，服务端维护 ContactsMap 的代码先屏蔽了
	// 如果要启用服务端判断，每条消息都通过 redis 是不行的，会崩，可以通过 goframe 在每个节点自己在内存中临时缓存一份
	// redis 保存的联系人 id 用于给 goframe 回源
	//key, expire := rediskey.ContactsMap(this.uid)
	//conn.Do(this.ctx, "SADD", key, req.UserId)
	//conn.Do(this.ctx, "EXPIRE", key, expire)
	return
}

// 删除联系人
func (this contacts) DeleteContacts(contactsId string) {
	// 我删他
	sql := "DELETE FROM contacts WHERE user_id = ? AND contacts_user_id = ?"
	this.db.Exec(sql, this.uid.ToString(), contactsId)
	sql = "UPDATE user SET contacts_version = contacts_version + 1 WHERE id = ?"
	this.db.Exec(sql, this.uid.ToString())
	caches.RebuildUserinfoCache(this.ctx, this.uid.ToString())
	// 他删我
	sql = "UPDATE contacts SET deleted = 1 WHERE user_id = ? AND contacts_user_id = ?"
	this.db.Exec(sql, contactsId, this.uid.ToString())
	sql = "UPDATE user SET contacts_version = contacts_version + 1 WHERE id = ?"
	this.db.Exec(sql, contactsId)
	caches.RebuildUserinfoCache(this.ctx, contactsId)
}

// 联系人列表
func (this contacts) ContactsList() (ret dto.ContactsListRes, err error) {
	sql := "SELECT t.deleted, u.id as UserId, u.username, u.nickname, u.avatar, u.gender " +
		" FROM contacts t LEFT JOIN user u ON t.contacts_user_id = u.id " +
		" WHERE t.user_id = ?"
	var list []dto.ContactsItem
	this.db.Raw(sql, this.uid.ToString()).Scan(&list)
	if len(list) > 0 {
		ret = dto.ContactsListRes{List: list}
	}
	return
}
