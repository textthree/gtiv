package webrtc

import (
	socketio "github.com/googollee/go-socket.io"
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/providers/goodlog"
)

func room(server *socketio.Server) {
	// 连接状态
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		goodlog.Info("[OnConnect] socketId: ", s.ID())
		return nil
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		goodlog.Error("[OnError]: ", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		goodlog.Info("[OnDisconnect]: ", reason)
	})

	// 用户加入房间，分布式环境下需要将房间信息保存到 redis
	// Golang 版本没有获取人数的 dto，要想控制人数只能使用 redis 自己保存
	server.OnEvent("/", "join", func(socket socketio.Conn, roomId string) {
		server.JoinRoom("/", roomId, socket)
		users := server.RoomLen("/", roomId)
		goodlog.Trace("用户加入房间 " + roomId + ", socketId: " + socket.ID() + ", 当前人数: " + strkit.Tostring(users))
		if users < 3 {
			// 返回给客户端告知进入房间成功，然后客户端进行音视频流绑定
			socket.Emit("joined", roomId, socket.ID())
			if users > 1 {
				// 通知房间里所有用户，有人加入，golalng 会发给房间所有用户，包括自己，
				// 因此客户端需要判断是否是自己发送的房间消息，如果是则忽略
				server.BroadcastToRoom("/", roomId, "otherjoin", socket.ID())
			}
		} else {
			// 如果房间里人满了
			socket.Leave(roomId)
			socket.Emit("full", roomId, socket.ID())
		}
	})

	// 离开房间
	server.OnEvent("/", "leave", func(socket socketio.Conn, roomId string) {
		server.LeaveRoom("/", roomId, socket)
		// 通知其他用户有人离开
		server.BroadcastToRoom("/", roomId, "bye", roomId, socket.ID())
		// 通知用户服务器已处理
		socket.Emit("leaved", roomId, socket.ID())
	})

	// 断开链接
	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", "bye "+last)
		goodlog.Yellow("关闭连接 " + last)
		s.Close()
		return last
	})

	// 中转消息
	server.OnEvent("/", "message", func(socket socketio.Conn, roomId, data string) {
		//roomId := "room-1"
		goodlog.Trace("客户端 " + socket.ID() + " 发送 " + roomId + " 房间消息: " + data)
		server.BroadcastToRoom("/", roomId, "message", socket.ID(), data)
	})

}
