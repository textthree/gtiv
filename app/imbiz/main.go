package main

import (
	"fmt"
	"github.com/text3cn/goodle/goodle"
	"github.com/text3cn/goodle/providers/etcd"
	"github.com/text3cn/goodle/providers/goodlog"
	"github.com/text3cn/goodle/providers/orm"
	"google.golang.org/grpc"
	_ "gtiv/app/imbiz/boot"
	"gtiv/app/imbiz/im"
	"gtiv/app/imbiz/internal"
	"gtiv/app/imbiz/internal/entity"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// @title imbiz API
// @description The Newframe IM
// @BasePath /
func main() {
	go goodle.Init(func() {
		//autoMigrate()
		//genapi()
	}).RunHttp(internal.Router)
	svr, cfg := im.New()
	etcd.Instance().ServiceRegister()
	rpcSrv := im.NewGrpcServer(cfg.RPCServer, svr)
	exitSignal(svr, rpcSrv)
}

func autoMigrate() {
	db := orm.GetDB()
	err := db.AutoMigrate(
		&entity.User{},
		&entity.Room{},
		&entity.UserRoom{},
		&entity.Faq{},
		&entity.Country{},
		&entity.Contacts{},
	)
	if err == nil {
		goodlog.Pink("migrate ok")
	}
}

func genapi() {
	command := exec.Command("/Users/t3/go/bin/swag", "init")
	output, err := command.Output()
	if err != nil {
		goodlog.Error("swag init error:", err)
	}
	fmt.Println(string(output))
}

func exitSignal(svr *im.Logic, rpcSrv *grpc.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		goodlog.Trace("imbiz get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			svr.Close()
			rpcSrv.GracefulStop()
			etcd.Instance().ServiceOffline()
			goodlog.Trace("imbiz [version: %s] exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
