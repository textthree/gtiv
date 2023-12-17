// 依赖库
var https = require('https');
var express = require('express');
var serveIndex = require('serve-index');
var fs = require('fs');
const {Server} = require("socket.io");

// 一个房间里可以同时在线的最大用户数
var USERCOUNT = 3;

var app = express();
// app.use(serveIndex('./web')); // 遍历目录作为首页
app.use(express.static(__dirname + "/client", {index: "index.html"}));
// 设置跨域访问
app.all("*", function (req, res, next) {
  // 设置允许跨域的域名， * 代表允许任意域名跨域
  res.header("Access-Control-Allow-Origin", "*");
  // 允许的header 类型
  res.header("Access-Control-Allow-Headers", "content-type");
  // 跨域允许的请求方式
  res.header("Access-Control-Allow-Methods", "DELETE,PUT,POST,GET,OPTIONS");
  if (req.method.toLowerCase() == 'options') {
    res.send(200); // 让options 尝试请求快速结束
  } else {
    next();
  }
});

// HTTP 服务
var options = {
  key: fs.readFileSync('./certificate/webrtc.text3.cn.key'),
  cert: fs.readFileSync('./certificate/webrtc.text3.cn_bundle.pem')
}
var server = https.createServer(options, app);

// socket 信令服务
const io = new Server(server);
io.sockets.on('connection', (socket) => {
  // 用户加入房间
  socket.on('join', (room) => {
    socket.join(room);
    // https://socket.io/docs/v4/rooms/
    // var myRoom = io.of("/").adapter.rooms.get(room);
    var myRoom = io.sockets.adapter.rooms.get(room);
    var users = (myRoom) ? myRoom.size : 0;
    console.log('用户加入房间 ' + room + '，socket.id:' + socket.id + '，当前人数：' + users);
    // 如果房间里人未满
    if (users < USERCOUNT) {
      // 返回给客户端告知进入房间成功，然后客户端进行音视频流绑定
      socket.emit('joined', room, socket.id);
      // 通知房间里所有用户，有人来了
      if (users > 1) {
        socket.to(room).emit('otherjoin', room, socket.id);
      }
    } else { // 如果房间里人满了
      socket.leave(room);
      socket.emit('full', room, socket.id);
    }
  });

  // 用户离开房间
  socket.on('leave', (room) => {

    // 从管理列表中将用户删除
    socket.leave(room);

    var myRoom = io.sockets.adapter.rooms[room];
    var users = (myRoom) ? Object.keys(myRoom.sockets).length : 0;
    console.debug('the user number of room is: ' + users);

    // 通知其他用户有人离开了
    socket.to(room).emit('bye', room, socket.id);

    // 通知用户服务器已处理
    socket.emit('leaved', room, socket.id);

  });

  // 中转消息
  socket.on('message', (room, data) => {
    console.debug('[中转消息] room: ' + room + "，type:" + data.type);
    socket.to(room).emit('message', room, data);
  });

});

server.listen(443, '172.31.26.174');
//server.listen(8008, '192.168.1.200');


console.log('启动成功')
