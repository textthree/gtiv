package im

import (
	"context"
	"github.com/text3cn/goodle/providers/goodlog"
	"gtiv/app/imbiz/im/conf"
	"gtiv/app/imbiz/im/dao"
)

var ImLogic *Logic

// Logic struct
type Logic struct {
	c   *conf.Config
	dao *dao.Dao
	// online
	totalIPs   int64
	totalConns int64
	roomCount  map[string]int32
}

// New init
func New() (l *Logic, config *conf.Config) {
	cfg := conf.InitConfig()
	ImLogic = &Logic{
		c:   cfg,
		dao: dao.New(cfg),
	}
	return ImLogic, cfg
}

// Ping ping resources is ok.
func (l *Logic) Ping(c context.Context) (err error) {
	return l.dao.Ping(c)
}

// Close close resources.
func (l *Logic) Close() {
	goodlog.Pink("Logic Close")
	//l.dao.Close()
}
