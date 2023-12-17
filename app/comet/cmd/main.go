package main

import (
	log "github.com/golang/glog"
	"github.com/text3cn/goodle/goodle"
	"github.com/text3cn/goodle/providers/etcd"
	"gtiv/app/comet"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// 私聊是能拿到 comet 的地址直接推送到指定 comet 的，不用广播的
// 注意 redis 能不能拿到 comet 的 ip 地址，原乡好像是 logic 去 discovery 中取的 comet 地址
// 原先是在 conf.go 中初始化环境变量获取本地 ip 保存到 discovery
// defHost, _    = os.Hostname()
// defAddrs      = os.Getenv("ADDRS") // 127.0.0.1 ,  comet 应该上报自己的
// defWeight, _  = strconv.ParseInt(os.Getenv("WEIGHT"), 10, 32)
// defOffline, _ = strconv.ParseBool(os.Getenv("OFFLINE"))
// defDebug, _   = strconv.ParseBool(os.Getenv("DEBUG"))
func main() {
	goodle.Init()
	if err := comet.Init(); err != nil {
		panic(err)
	}
	// new comet server
	srv := comet.NewServer(comet.Conf)
	if err := comet.InitWhitelist(comet.Conf.Whitelist); err != nil {
		panic(err)
	}
	logicCputNum := runtime.NumCPU() // 8 核 16 线程的 cpu 值为 16
	if err := comet.InitTCP(srv, comet.Conf.TCP.Bind, logicCputNum); err != nil {
		panic(err)
	}
	if err := comet.InitWebsocket(srv, comet.Conf.Websocket.Bind, logicCputNum); err != nil {
		panic(err)
	}
	if comet.Conf.Websocket.TLSOpen {
		if err := comet.InitWebsocketWithTLS(srv, comet.Conf.Websocket.TLSBind, comet.Conf.Websocket.CertFile, comet.Conf.Websocket.PrivateFile, runtime.NumCPU()); err != nil {
			panic(err)
		}
	}
	etcd.Instance().ServiceRegister()
	// new grpc server
	rpcSrv := comet.New(comet.Conf.RPCServer, srv)
	// 发现 logic

	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-comet get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			rpcSrv.GracefulStop()
			srv.Close()
			etcd.Instance().ServiceOffline()
			log.Infof("goim-comet [version: %s] exit")
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
