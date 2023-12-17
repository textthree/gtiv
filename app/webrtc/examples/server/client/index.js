// 本地视频预览窗口
var localVideo = document.querySelector('video#localvideo ');
// 远端视频预览窗口
var remoteVideo = document.querySelector('video#remotevideo ');
// 连接信令服务器 Button
var btnConn = document.querySelector('button#connserver ');
// 与信令服务器断开连接Button
var btnLeave = document.querySelector('button#leave ');
// 查看Offer 文本窗口
var offer = document.querySelector('textarea#offer ');
// 查看Answer 文本窗口
var answer = document.querySelector('textarea#answer ');
// 本地视频流
var localStream = null;
// 远端视频流
var remoteStream = null;
// PeerConnection
var pc = null;
// 房间号
var roomid;
var socket = null;
// offer 描述
var offerdesc = null;
// 状态机，初始为init
var state = 'init';

// coturn
var connConfig2 = {
  'iceServers': [
    {
      'urls': 'stun:192.168.1.200:80',
    },
    {
      'urls': 'turn:192.168.1.200:80',
      'username': "your_username",
      'credential': "your_password"
    },
    {url: 'stun:stun.l.google.com:19302'}
  ],
  // "iceTransportPolicy": "relay", // 强制使用 turn 传输，用来做测试用，能 p2p 就不用服务器转发
};

// pion
var connConfig = {
  'iceServers':
    [
      {
        'urls': 'stun:192.168.1.200:5800',
      },
      {
        'urls': 'turn:192.168.1.200:5800',
        'username': "foo",
        'credential': "var"
      }
    ],
  "iceTransportPolicy": "relay", // 强制使用 turn 传输，用来做测试用，能 p2p 就不用服务器转发
};


// 响应连接按钮，打开音视频设备并连接信令服务器
function connSignalServer() {
  if (!navigator.mediaDevices ||
    !navigator.mediaDevices.getUserMedia) {
    console.error('the getUserMedia is not supported!');
    return;
  } else {
    var constraints;
    constraints = {
      video: true,
      audio: {
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: true
      }
    }
    navigator.mediaDevices.getUserMedia(constraints)
      .then(getLocalMediaStream)
      .catch(handleError);
  }
  return true;
}

// 打开音视频设备成功时的回调函数
function getLocalMediaStream(stream) {
  // 将从设备上获取到的音视频 track 添加到 localStream 中
  if (localStream) {
    stream.getAudioTracks().forEach((track) => {
      localStream.addTrack(track);
      stream.removeTrack(track);
    });
  } else {
    localStream = stream;
  }
  // 本地视频标签与本地流绑定
  localVideo.srcObject = localStream;
  // 调用 conn() 函数的位置特别重要，一定要在 getLocalMediaStream() 调用之后再调用它，
  // 否则就会出现绑定失败的情况
  conn();
}

// 与信令服务器建立 socket.io 连接，并根据信令更新状态机。
function conn() {
  // 连接信令服务器
  socket = io.connect();

  // joined 消息处理函数
  socket.on('joined', (roomid, id) => {
    // 状态机变更为 joined
    state = 'joined'
    console.log('加入房间 ' + roomid + " 成功");

    // 如果是 Mesh 方案，第一个人不该在这里创建 peerConnection，
    // 而是要等到所有端都收到一个 otherjoin 消息时再创建

    // 创建 PeerConnection 并绑定音视频轨
    createPeerConnection();
    bindTracks();

    // 设置 button 状态
    btnConn.disabled = true;
    btnLeave.disabled = false;
  });

  // otherjoin 消息处理函数
  socket.on('otherjoin', (roomid) => {
    console.log('receive joined message:', roomid, state);
    // 如果是多人，每加入一个人都要创建一个新的 PeerConnection
    if (state === 'joined_unbind') {
      createPeerConnection();
      bindTracks();
    }
    // 状态机变更为 joined_conn
    state = 'joined_conn';
    // 开始呼叫对方
    call();
    console.log('receive other_join message, state = ', state);
  });

  //full 消息处理函数
  socket.on('full', (roomid, id) => {
    console.log('receive full message ', roomid, id);
    // 关闭 socket.io 连接
    socket.disconnect();
    // 挂断呼叫
    hangup();
    // 关闭本地媒体
    closeLocalMedia();
    // 状态机变更为leaved
    state = 'leaved';
    console.log('receive full message , state = ', state);
    alert('房间人数已满!');
  });

  // leaved 消息处理函数
  socket.on('leaved', (roomid, id) => {
    console.log('receive leaved message ', roomid, id);
    // 状态机变更为 leaved
    state = 'leave'
    // 关闭 socket.io 连接
    socket.disconnect();
    console.log('receive leaved message, state = ', state);
    // 改变button 状态
    btnConn.disabled = false;
    btnLeave.disabled = true;
  });

  // bye 消息处理函数
  socket.on('bye', (room, id) => {
    console.log('receive bye message ', roomid, id);

    // Mesh 方案时，应该带上当前房间的用户数，如果当前房间用户数不小于 2 则不用修改状态，并且关闭的应该是对应用户的 PeerConnection
    // 在客户端应该维护一张 PeerConnection 表， 它是 key:value 的格式， key=userid , value=peerconnection

    // 状态机变更为joined_unbind
    state = 'joined_unbind';
    // 挂断呼叫
    hangup();
    offer.value = '';
    answer.value = '';
    console.log('receive bye message, state=', state);
  });

  // socket.io 连接断开处理函数
  socket.on('disconnect', (socket) => {
    console.log('receive disconnect message!', roomid);
    if (!(state === 'leaved')) {
      // 挂断呼叫
      hangup();
      // 关闭本地媒体
      closeLocalMedia();
    }
    // 状态机变更为leaved
    state = 'leaved';
  });

  // 收到对端消息处理函数
  socket.on('message', (roomid, data) => {
    console.log("长连接来信（" + data.type + "）：", roomid, data);
    if (data === null || data === undefined) {
      console.error('the message is invalid!');
      return;
    }

    // 如果收到的 SDP 是 offer
    if (data.hasOwnProperty('type') && data.type === 'offer') {
      offer.value = data.sdp;
      // 进行媒体协商
      pc.setRemoteDescription(new RTCSessionDescription(data));
      // 创建 answer
      pc.createAnswer()
        .then(getAnswer)
        .catch(handleAnswerError);
    } else if (data.hasOwnProperty('type') && data.type == 'answer') {
      // 如果收到的 SDP 是 answer
      answer.value = data.sdp;
      // 进行媒体协商
      pc.setRemoteDescription(new RTCSessionDescription(data));
      // 如果收到的是 Candidate 消息
    } else if (data.hasOwnProperty('type') && data.type === 'candidate') {
      var candidate = new RTCIceCandidate({
        sdpMLineIndex: data.label,
        candidate: data.candidate
      });
      // 将远端 Candidate 消息添加到 PeerConnection 中
      pc.addIceCandidate(candidate);
    } else {
      console.log('无效的信令指令!', data);
    }
  });

  // 从 url 中获取 roomid
  roomid = getQueryVariable('room');
  roomid = "room-1"
  // 发送 join 消息
  socket.emit('join', roomid);
  return true;
}


// 使用 socket 发送消息
function sendMessage(roomid, data) {
  console.log('发送消息到房间，消息内容：', roomid, data);
  if (!socket) {
    console.log('socket is null');
  }
  socket.emit('message', roomid, data);
}

// 获取远端媒体流
function getRemoteStream(e) {
  console.log("[获取远端媒体流]", e.streams[0])
  // 存放远端视频流
  remoteStream = e.streams[0];
  // 远端视频标签与远端视频流绑定
  remoteVideo.srcObject = e.streams[0];
}


// 获取 Answer SDP 描述符的回调函数
function getAnswer(desc) {
  // 设置 Answer
  pc.setLocalDescription(desc);
  // 将 Answer 显示出来
  answer.value = desc.sdp;
  // 将 Answer SDP 发送给对端
  sendMessage(roomid, desc);
}

// 获取Offer SDP 描述符的回调函数
function getOffer(desc) {
  // 设置 Offer
  pc.setLocalDescription(desc);
  // 将Offer 显示出来
  offer.value = desc.sdp;
  offerdesc = desc;
  // 将Offer SDP 发送给对端
  sendMessage(roomid, offerdesc);
}

// 创建PeerConnection 对象
function createPeerConnection() {

  // 如果是多人的话，在这里要创建一个新的连接，新创建好的要放到一个映射表中
  // key=userid , value=peerconnection

  if (!pc) {
    // 创建 PeerConnection 对象
    pc = new RTCPeerConnection(connConfig);
    // 当收集到 Candidate 后
    pc.onicecandidate = (e) => {
      if (e.candidate) {
        console.log("candidate: " + JSON.stringify(e.candidate.toJSON()));
        // 将 Candidate 发送给对端
        sendMessage(roomid, {
          type: 'candidate',
          label: event.candidate.sdpMLineIndex,
          id: event.candidate.sdpMid,
          candidate: event.candidate.candidate
        });
      } else {
        console.log('this is the end candidate');
      }
    }
    // 当 PeerConnection 对象收到远端音视频流时，触发 ontrack 事件，并回调 getRemoteStream 函数
    pc.ontrack = getRemoteStream;
  } else {
    console.error('[ERROR]: the peerConnection have be created!');
  }
  console.log('Create RTCPeerConnection success!');
  return;
}

// 将音视频 track 绑定到 PeerConnection 对象中
function bindTracks() {
  if (pc === null && localStream === undefined) {
    console.error('[ERROR]: PeerConnection is null or undefined!');
    return;
  }
  if (localStream === null && localStream === undefined) {
    console.error('[ERROR]: localstream is null or undefined!');
    return;
  }
  // 将本地音视频流中所有的track 添加到PeerConnection 对象中
  localStream.getTracks().forEach((track) => {
    pc.addTrack(track, localStream);
  });
  console.log('Bind tracks into RTCPeerConnection!');
}

/**
 547 * 功能： 开启“ 呼叫”
 548 *
 549 * 返回值： 无
 550 */
function call() {

  if (state === 'joined_conn') {

    var offerOptions = {
      offerToReceiveAudio: 1,
      offerToReceiveVideo: 1
    }

    /**
     561           * 创建Offer ，
     562           * 如果成功， 则回调getOffer () 方法
     563           * 如果失败， 则回调handleOfferError () 方法
     564      */
    pc.createOffer(offerOptions)
      .then(getOffer)
      .catch(handleOfferError);
  }
}

/**
 572 * 功能： 挂断“ 呼叫”
 573 *
 574 * 返回值： 无
 575 */
function hangup() {

  if (!pc) {
    return;
  }

  offerdesc = null;

  // 将PeerConnection 连接关掉
  pc.close();
  pc = null;

}

/**
 591 * 功能： 关闭本地媒体
 592 *
 593 * 返回值： 无
 594 */
function closeLocalMedia() {

  if (!(localStream === null ||
    localStream === undefined)) {
    // 遍历每个track ， 并将其关闭
    localStream.getTracks().forEach((track) => {
      track.stop();
    });
  }
  localStream = null;
}

/**
 608 * 功能： 离开房间
 609 *
 610 * 返回值： 无
 611 */
function leave() {

  // 向信令服务器发送leave 消息
  socket.emit('leave', roomid);

  // 挂断“ 呼叫”
  hangup();
  // 关闭本地媒体
  closeLocalMedia();

  offer.value = '';
  answer.value = '';
  btnConn.disabled = false;
  btnLeave.disabled = true;
}

// 为Button 设置单击事件
btnConn.onclick = connSignalServer
btnLeave.onclick = leave;
