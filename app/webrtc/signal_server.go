package webrtc

import (
	socketio "github.com/googollee/go-socket.io"
	"github.com/text3cn/goodle/providers/goodlog"
	"net/http"
)

// https://github.com/googollee/go-socket.io/tree/master/_examples
func StartSingnalServer() {

	server := socketio.NewServer(nil)

	// 两个客户端加入的是同一台服务器，Redis 不会生效
	_, err := server.Adapter(&socketio.RedisAdapterOptions{
		Addr:     redisAddr,
		Password: redisPwd,
		DB:       redisDb,
		Prefix:   "socket.io",
	})
	if err != nil {
		goodlog.Error("Redis adapter error: ", err)
	}

	one2one(server)
	// room(server)

	go server.Serve()
	defer server.Close()

	// 根命名空间必须监听一个，否则客户端连接不上
	server.OnEvent("/", "test", func(s socketio.Conn, msg string) {})

	// 心跳轮询接口，允许跨域
	// http.Handle("/socket.io/", server)
	http.HandleFunc("/socket.io/", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		goodlog.Info("轮询来自客户端：", origin)
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE,PUT,POST,GET,OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		server.ServeHTTP(w, r)
	})

	// 静态资源服务，用于访问 http://localhost:8008/index.html
	http.Handle("/", http.FileServer(http.Dir("./signalserver")))

	goodlog.Info("Socket.io 启动服务监听" + signalListenAddress)
	goodlog.Fatal(http.ListenAndServe(signalListenAddress, nil))
}
