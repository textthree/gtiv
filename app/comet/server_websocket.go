package comet

import (
	"context"
	"crypto/tls"
	"gtiv/kit/impkg/bytes"
	"gtiv/kit/impkg/logger"
	"gtiv/kit/impkg/protocol"
	xtime "gtiv/kit/impkg/time"
	"gtiv/kit/impkg/websocket"
	"io"
	"net"
	"strings"
	"time"

	log "github.com/golang/glog"
)

// InitWebsocket listen all tcp.bind and start accept connections.
func InitWebsocket(server *Server, addrs []string, logicCpuNum int) (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
			log.Errorf("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			log.Errorf("net.ListenTCP(tcp, %s) error(%v)", bind, err)
			return
		}
		log.Infof("start ws listen: %s", bind)
		// split N core accept
		for i := 0; i < logicCpuNum; i++ {
			go acceptWebsocket(server, listener)
		}
	}
	return
}

// InitWebsocketWithTLS init websocket with tls.
func InitWebsocketWithTLS(server *Server, addrs []string, certFile, privateFile string, accept int) (err error) {
	var (
		bind     string
		listener net.Listener
		cert     tls.Certificate
		certs    []tls.Certificate
	)
	certFiles := strings.Split(certFile, ",")
	privateFiles := strings.Split(privateFile, ",")
	for i := range certFiles {
		cert, err = tls.LoadX509KeyPair(certFiles[i], privateFiles[i])
		if err != nil {
			log.Errorf("Error loading certificate. error(%v)", err)
			return
		}
		certs = append(certs, cert)
	}
	tlsCfg := &tls.Config{Certificates: certs}
	tlsCfg.BuildNameToCertificate()
	for _, bind = range addrs {
		if listener, err = tls.Listen("tcp", bind, tlsCfg); err != nil {
			log.Errorf("net.ListenTCP(tcp, %s) error(%v)", bind, err)
			return
		}
		log.Infof("start wss listen: %s", bind)
		// split N core accept
		for i := 0; i < accept; i++ {
			go acceptWebsocketWithTLS(server, listener)
		}
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptWebsocket(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)

	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			logger.Cred("listener.Accept(%s) error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(server.c.TCP.KeepAlive); err != nil {
			logger.Cred("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(server.c.TCP.Rcvbuf); err != nil {
			logger.Cred("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(server.c.TCP.Sndbuf); err != nil {
			logger.Cred("conn.SetWriteBuffer() error(%v)", err)
			return
		}
		server.ServeWebsocket(conn, r)
		if r++; r == maxInt {
			// 可以通过 maxInt 限制单个 comet 实例的连接人数
			r = 0
		}
		logger.Cgreen("连接成功")
	}
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptWebsocketWithTLS(server *Server, lis net.Listener) {
	var (
		conn net.Conn
		err  error
		r    int
	)
	for {
		if conn, err = lis.Accept(); err != nil {
			// if listener close then return
			log.Errorf("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		server.ServeWebsocket(conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

// 响应客户端连接
func (s *Server) ServeWebsocket(conn net.Conn, r int) {
	if Conf.Debug {
		localAddr := conn.LocalAddr().String()
		remoteAddr := conn.RemoteAddr().String()
		logger.Info("Client ", remoteAddr, " dial to ", localAddr)
	}
	var (
		tr      = s.round.Timer(r)
		rp      = s.round.Reader(r)
		wp      = s.round.Writer(r)
		err     error
		rid     string
		accepts []int32
		hb      int64
		white   bool
		p       *protocol.Proto
		b       *Bucket
		trd     *xtime.TimerData
		lastHB  = time.Now()
		rb      = rp.Get()
		ch      = NewChannel(s.c.Protocol.CliProto, s.c.Protocol.SvrProto)
		rr      = &ch.Reader
		wr      = &ch.Writer
		ws      *websocket.Conn // websocket
		req     *websocket.Request
	)
	// reader
	ch.Reader.ResetBuffer(conn, rb.Bytes())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 握手
	step := 0
	trd = tr.Add(time.Duration(s.c.Protocol.HandshakeTimeout), func() {
		// NOTE: fix close block for tls
		_ = conn.SetDeadline(time.Now().Add(time.Millisecond * 100))
		_ = conn.Close()
		logger.Cred("[Websockt handshake timeout]", " key:", ch.Key, " remoteIP:", conn.RemoteAddr().String(), " step:", step)
	})
	// websocket
	ch.IP, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	step = 1
	if req, err = websocket.ReadRequest(rr); err != nil || req.RequestURI != "/sub" {
		conn.Close()
		tr.Del(trd)
		rp.Put(rb)
		if err != io.EOF {
			logger.Cred("[http.ReadRequest(rr) error]", err)
		}
		return
	}
	// writer
	wb := wp.Get()
	ch.Writer.ResetBuffer(conn, wb.Bytes())
	step = 2
	if ws, err = websocket.Upgrade(conn, rr, wr, req); err != nil {
		conn.Close()
		tr.Del(trd)
		rp.Put(rb)
		wp.Put(wb)
		if err != io.EOF {
			logger.Cred("[Websocket.NewServerConn error]", err)
		}
		return
	}
	// must not setadv, only used in auth
	step = 3
	if p, err = ch.CliProto.Set(); err == nil {
		if ch.Mid, ch.Key, rid, accepts, hb, err = s.authWebsocket(ctx, ws, p, req.Header.Get("Cookie")); err == nil {
			ch.Watch(accepts...) // 监听房间列表
			b = s.Bucket(ch.Key) // 根据用户 key 选择一个 bucket（对 key 做 cityhash 在取模），这个 key 是在 logic 模块返回的
			err = b.Put(rid, ch) // 将用户 id 和连接 channel 维护到 bucket 中
			if Conf.Debug {
				logger.Cgreen("[Websocket connected]", " key:", ch.Key, " mid:", ch.Mid, " proto:", p)
			}

		}
	}
	step = 4
	if err != nil {
		ws.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		if err != io.EOF && err != websocket.ErrMessageClose {
			logger.Cred("[Websocket handshake failed] ", "key:", ch.Key, " remoteIP:", conn.RemoteAddr().String(), " step:", step, "\nError:", err.Error())
		}
		return
	}
	trd.Key = ch.Key
	tr.Set(trd, hb)
	white = whitelist.Contains(ch.Mid)
	if white {
		whitelist.Printf("key: %s[%s] auth\n", ch.Key, rid)
		logger.Info("[whitelist]", " key:", ch.Key, " auth:", rid)
	}
	// handshake ok start dispatch goroutine
	step = 5
	go s.dispatchWebsocket(ws, wp, wb, ch)
	serverHeartbeat := s.RandServerHearbeat()
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		if white {
			whitelist.Printf("key: %s start read proto\n", ch.Key)
			logger.Info("[whitelist2]")
		}
		if err = p.ReadWebsocket(ws); err != nil {
			break
		}
		if white {
			logger.Info("[whitelist3]")
			whitelist.Printf("key: %s read proto:%v\n", ch.Key, p)
		}
		if p.Op == protocol.OpHeartbeat {
			tr.Set(trd, hb)
			p.Op = protocol.OpHeartbeatReply
			p.Body = nil
			// NOTE: send server heartbeat for a long time
			if now := time.Now(); now.Sub(lastHB) > serverHeartbeat {
				if err1 := s.Heartbeat(ctx, ch.Mid, ch.Key); err1 == nil {
					lastHB = now
				}
			}
			if Conf.Debug {
				logger.Info("[Websocket heartbeat receive]", " key:", ch.Key, " mid:", ch.Mid)
			}
			step++
		} else {
			if err = s.Operate(ctx, p, ch, b); err != nil {
				break
			}
		}
		if white {
			whitelist.Printf("key: %s process proto:%v\n", ch.Key, p)
		}
		ch.CliProto.SetAdv()
		ch.Signal()
		if white {
			whitelist.Printf("key: %s signal\n", ch.Key)
		}
	}
	if white {
		whitelist.Printf("key: %s server tcp error(%v)\n", ch.Key, err)
	}
	if err != nil && err != io.EOF && err != websocket.ErrMessageClose && !strings.Contains(err.Error(), "closed") {
		log.Errorf("key: %s server ws failed error(%v)", ch.Key, err)
	}
	b.Del(ch)
	tr.Del(trd)
	ws.Close()
	ch.Close()
	rp.Put(rb)
	if err = s.Disconnect(ctx, ch.Mid, ch.Key); err != nil {
		log.Errorf("key: %s operator do disconnect error(%v)", ch.Key, err)
	}
	if white {
		whitelist.Printf("key: %s disconnect error(%v)\n", ch.Key, err)
	}
	if Conf.Debug {
		log.Infof("websocket disconnected key: %s mid:%d", ch.Key, ch.Mid)
	}
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (s *Server) dispatchWebsocket(ws *websocket.Conn, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	var (
		err    error
		finish bool
		online int32
		white  = whitelist.Contains(ch.Mid)
	)
	if Conf.Debug {
		log.Infof("key: %s start dispatch tcp goroutine", ch.Key)
	}
	for {
		if white {
			whitelist.Printf("key: %s wait proto ready\n", ch.Key)
		}
		var p = ch.Ready()
		if white {
			whitelist.Printf("key: %s proto ready\n", ch.Key)
		}
		if Conf.Debug {
			log.Infof("key:%s dispatch msg:%s", ch.Key, p.Body)
		}
		switch p {
		case protocol.ProtoFinish:
			if white {
				whitelist.Printf("key: %s receive proto finish\n", ch.Key)
			}
			if Conf.Debug {
				log.Infof("key: %s wakeup exit dispatch goroutine", ch.Key)
			}
			finish = true
			goto failed
		case protocol.ProtoReady:
			// fetch message from svrbox(client send)
			for {
				if p, err = ch.CliProto.Get(); err != nil {
					break
				}
				if white {
					whitelist.Printf("key: %s start write client proto%v\n", ch.Key, p)
				}
				if p.Op == protocol.OpHeartbeatReply {
					if ch.Room != nil {
						online = ch.Room.OnlineNum()
					}
					if err = p.WriteWebsocketHeart(ws, online); err != nil {
						goto failed
					}
				} else {
					if err = p.WriteWebsocket(ws); err != nil {
						goto failed
					}
				}
				if white {
					whitelist.Printf("key: %s write client proto%v\n", ch.Key, p)
				}
				p.Body = nil // avoid memory leak
				ch.CliProto.GetAdv()
			}
		default:
			if white {
				whitelist.Printf("key: %s start write server proto%v\n", ch.Key, p)
			}
			if err = p.WriteWebsocket(ws); err != nil {
				goto failed
			}
			if white {
				whitelist.Printf("key: %s write server proto%v\n", ch.Key, p)
			}
			if Conf.Debug {
				log.Infof("websocket sent a message key:%s mid:%d proto:%+v", ch.Key, ch.Mid, p)
			}
		}
		if white {
			whitelist.Printf("key: %s start flush \n", ch.Key)
		}
		// only hungry flush response
		if err = ws.Flush(); err != nil {
			break
		}
		if white {
			whitelist.Printf("key: %s flush\n", ch.Key)
		}
	}
failed:
	if white {
		whitelist.Printf("key: %s dispatch tcp error(%v)\n", ch.Key, err)
	}
	if err != nil && err != io.EOF && err != websocket.ErrMessageClose {
		log.Errorf("key: %s dispatch ws error(%v)", ch.Key, err)
	}
	ws.Close()
	wp.Put(wb)
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = (ch.Ready() == protocol.ProtoFinish)
	}
	if Conf.Debug {
		log.Infof("key: %s dispatch goroutine exit", ch.Key)
	}
}

// auth for goim handshake with client, use rsa & aes.
func (s *Server) authWebsocket(ctx context.Context, ws *websocket.Conn, p *protocol.Proto, cookie string) (mid int64, key, rid string, accepts []int32, hb int64, err error) {
	for {
		if err = p.ReadWebsocket(ws); err != nil {
			return
		}
		if p.Op == protocol.OpAuth {
			break
		} else {
			logger.Cred("ws request operation ", p.Op, "  not auth")
		}
	}
	if mid, key, rid, accepts, hb, err = s.Connect(ctx, p, cookie); err != nil {
		return
	}
	p.Op = protocol.OpAuthReply
	p.Body = nil
	if err = p.WriteWebsocket(ws); err != nil {
		return
	}
	err = ws.Flush()
	return
}
