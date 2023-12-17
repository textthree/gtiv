package webrtc

import (
	socketio "github.com/googollee/go-socket.io"
	"github.com/text3cn/goodle/providers/goodlog"
)

var namespace = "/one2one"

func one2one(server *socketio.Server) {
	one2oneConnState(server)

	// 加入房间，返回 socketId 给客户端
	server.OnEvent(namespace, "join", func(socket socketio.Conn, roomId string) {
		server.JoinRoom(namespace, roomId, socket)
		//logger.Cgreen("客户端 " + socket.ID() + " 加入房间 " + roomId)
		socket.Emit("joined", socket.ID())
	})

	// 中转消息
	server.OnEvent(namespace, "message", func(socket socketio.Conn, roomId, msg string) {
		//logger.Cpink("客户端 " + socket.ID() + " 发送消息到房间 " + roomId + "\n" + "消息内容：" + msg)
		server.BroadcastToRoom(namespace, roomId, "message", socket.ID(), msg)
	})
}

func one2oneConnState(server *socketio.Server) {
	// 连接状态
	server.OnConnect(namespace, func(socket socketio.Conn) error {
		goodlog.Trace("[OnConnect] socketId: ", socket.ID())
		return nil
	})
	server.OnError(namespace, func(s socketio.Conn, e error) {
		goodlog.Error("[OnError]: ", e)
	})
	server.OnDisconnect(namespace, func(socket socketio.Conn, reason string) {
		goodlog.Trace("[OnDisconnect]: ", socket.ID()+" leave room. "+reason)
	})
}
