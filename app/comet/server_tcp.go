package comet

import (
	"context"
	"fmt"
	"gtiv/kit/impkg/logger"
	"io"
	"net"
	"strings"

	"gtiv/kit/impkg/bufio"
	"gtiv/kit/impkg/bytes"
	"gtiv/kit/impkg/protocol"
	xtime "gtiv/kit/impkg/time"
	"time"
)

const (
	maxInt = 1<<31 - 1
)

// 在 cmd/comet/main.go 中启动 TCP 连接服务，接收客户端的连接
func InitTCP(server *Server, addrs []string, accept int) (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
			logger.Error("net.ResolveTCPAddr(tcp, %s) error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			logger.Error("net.ListenTCP(tcp, %s) error(%v)", bind, err)
			return
		}
		logger.Infof("启动 Comet Tcp 服务，监听 %s 端口", bind)
		// 为每个连接分配一个协程
		for i := 0; i < accept; i++ {
			go acceptTCP(server, listener)
		}
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptTCP(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			logger.Error("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(server.c.TCP.KeepAlive); err != nil {
			logger.Error("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(server.c.TCP.Rcvbuf); err != nil {
			logger.Error("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(server.c.TCP.Sndbuf); err != nil {
			logger.Error("conn.SetWriteBuffer() error(%v)", err)
			return
		}
		go serveTCP(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serveTCP(s *Server, conn *net.TCPConn, r int) {
	var (
		// timer
		tr = s.round.Timer(r)
		rp = s.round.Reader(r)
		wp = s.round.Writer(r)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	if Conf.Debug {
		logger.Cgreen("[新连接上线]", " 从 ", rAddr, " 接入 ", lAddr)
	}
	s.ServeTCP(conn, rp, wp, tr)
}

// 干活的 worker，用于对客户端收发消息
func (s *Server) ServeTCP(conn *net.TCPConn, rp, wp *bytes.Pool, tr *xtime.Timer) {
	var (
		err     error
		roomId  string
		accepts []int32
		hb      int64
		white   bool
		p       *protocol.Proto
		bucket  *Bucket
		trd     *xtime.TimerData
		lastHb  = time.Now()
		rb      = rp.Get()
		wb      = wp.Get()
		ch      = NewChannel(s.c.Protocol.CliProto, s.c.Protocol.SvrProto)
		rr      = &ch.Reader
		wr      = &ch.Writer
	)
	ch.Reader.ResetBuffer(conn, rb.Bytes())
	ch.Writer.ResetBuffer(conn, wb.Bytes())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 握手
	step := 0
	trd = tr.Add(time.Duration(s.c.Protocol.HandshakeTimeout), func() {
		// NOTE: fix close block for tls
		//_ = conn.SetDeadline(time.Now().Add(time.Millisecond * 100))
		conn.Close()
		logger.Cred("[连接过期] key: %s remoteIP: %s step: %d", ch.Key, conn.RemoteAddr().String(), step)
	})
	ch.IP, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	// must not setadv, only used in auth
	step = 1

	if p, err = ch.CliProto.Set(); err == nil {
		if ch.Mid, ch.Key, roomId, accepts, hb, err = s.authTCP(ctx, rr, wr, p); err == nil {
			ch.Watch(accepts...)
			bucket = s.Bucket(ch.Key)    // 为连接分配 bucket
			err = bucket.Put(roomId, ch) // 将连接与 room 关联
			if Conf.Debug {
				logger.Infof("[连接成功] tcp connnected key:%s mid:%d room:%s proto:%+v", ch.Key, ch.Mid, roomId, p)
			}
			if err != nil {
				logger.Cred("出错了", err)
			}
		}
	}
	step = 2
	if err != nil {
		conn.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		logger.Error("握手失败 ", ch.Key, err)
		return
	}
	trd.Key = ch.Key
	tr.Set(trd, hb)
	white = whitelist.Contains(ch.Mid)
	if white {
		whitelist.Printf("key: %s[%s] auth\n", ch.Key, roomId)
	}
	step = 3
	// 握手成功，分派 goroutine 干活，然后主线程 for 换中不停监听等 job 发来消息塞进管道
	// dispatchTCP() 从管道中读取消息，写给客户端
	go s.dispatchTCP(conn, wr, wp, wb, ch)
	serverHeartbeat := s.RandServerHearbeat()
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			logger.Error("ch.CliProto.Set() Error:", err)
			break
		}
		if white {
			whitelist.Printf("key: %s start read proto\n", ch.Key)
		}
		if err = p.ReadTCP(rr); err != nil {
			break
		}
		if white {
			whitelist.Printf("key: %s read proto:%v\n", ch.Key, p)
		}
		if p.Op == protocol.OpHeartbeat {
			tr.Set(trd, hb)
			p.Op = protocol.OpHeartbeatReply
			p.Body = nil
			// NOTE: send server heartbeat for a long time
			if now := time.Now(); now.Sub(lastHb) > serverHeartbeat {
				if err1 := s.Heartbeat(ctx, ch.Mid, ch.Key); err1 == nil {
					lastHb = now
				}
			}
			if Conf.Debug {
				logger.Infof("tcp heartbeat receive key:%s, mid:%d", ch.Key, ch.Mid)
			}
			step++
		} else {
			if err = s.Operate(ctx, p, ch, bucket); err != nil {
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
	if err != nil && err != io.EOF && !strings.Contains(err.Error(), "closed") {
		logger.Error("key: %s server tcp failed error(%v)", ch.Key, err)
	}
	bucket.Del(ch)
	tr.Del(trd)
	rp.Put(rb)
	conn.Close()
	ch.Close()
	if err = s.Disconnect(ctx, ch.Mid, ch.Key); err != nil {
		logger.Error("key: %s mid: %d operator do disconnect error(%v)", ch.Key, ch.Mid, err)
	}
	if white {
		whitelist.Printf("key: %s mid: %d disconnect error(%v)\n", ch.Key, ch.Mid, err)
	}
	if Conf.Debug {
		logger.Infof("tcp disconnected key: %s mid: %d", ch.Key, ch.Mid)
	}
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (s *Server) dispatchTCP(conn *net.TCPConn, bufioWriter *bufio.Writer, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	var (
		err    error
		finish bool
		online int32
		white  = whitelist.Contains(ch.Mid)
	)
	if Conf.Debug {
		logger.Infof("key: %s start dispatch tcp goroutine", ch.Key)
	}
	for {
		if white {
			logger.Infof("key: %s wait proto ready\n", ch.Key)
		}
		var p = ch.Ready() // 从消息管道中读取消息
		if white {
			logger.Infof("key: %s proto ready\n", ch.Key)
		}
		if Conf.Debug {
			logger.Infof("key:%s dispatch msg:%v", ch.Key, *p)
		}
		switch p {
		case protocol.ProtoFinish:
			if white {
				logger.Infof("key: %s receive proto finish\n", ch.Key)
			}
			if Conf.Debug {
				logger.Infof("key: %s 唤醒退出派遣的 goroutine", ch.Key)
			}
			finish = true
			logger.Cred("goto1:")
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
					if err = p.WriteTCPHeart(bufioWriter, online); err != nil {
						logger.Cred("goto2:")
						goto failed
					}
				} else {
					if err = p.WriteTCP(bufioWriter); err != nil {
						logger.Cred("goto3:")
						goto failed
					}
				}
				if white {
					logger.Errorf("key: %s write client proto%v\n", ch.Key, p)
				}
				p.Body = nil // avoid memory leak
				ch.CliProto.GetAdv()
			}
		default:
			if white {
				logger.Errorf("key: %s start write server proto%v\n", ch.Key, p)
			}
			// 将消息发送给客户端
			if err = p.WriteTCP(bufioWriter); err != nil {
				logger.Cred("goto4:")
				goto failed
			}
			if white {
				logger.Errorf("key: %s write server proto%v\n", ch.Key, p)
			}
			if Conf.Debug {
				//logger.Infof("tcp sent a message key:%s mid:%d proto:%+v", ch.Key, ch.Mid, p)
			}
		}
		if white {
			logger.Errorf("key: %s start flush \n", ch.Key)
		}

		// only hungry flush response
		if err = bufioWriter.Flush(); err != nil {
			logger.Cred("Flush 出错:", err)
			break
		}
		if white {
			whitelist.Printf("key: %s flush\n", ch.Key)
		}
	}
failed:
	if white {
		logger.Errorf("key: %s dispatch tcp error(%v)\n", ch.Key, err)
	}
	if err != nil {
		logger.Error("key: %s dispatch tcp error(%v)", ch.Key, err)
	}
	conn.Close()
	wp.Put(wb)
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = (ch.Ready() == protocol.ProtoFinish)
	}
	if Conf.Debug {
		info := fmt.Errorf("key: %s dispatch goroutine exit", ch.Key)
		logger.Info(info)
	}
}

// auth for goim handshake with client, use rsa & aes.
func (s *Server) authTCP(ctx context.Context, rr *bufio.Reader, wr *bufio.Writer, p *protocol.Proto) (mid int64, key, roomId string, accepts []int32, hb int64, err error) {
	for {
		if err = p.ReadTCP(rr); err != nil {
			return
		}
		// 等待客户端发送鉴权请求，否则建立连接后就一直阻塞
		if p.Op == protocol.OpAuth {
			break
		} else {
			logger.Error("tcp request operation(%d) not auth", p.Op)
		}
	}
	if mid, key, roomId, accepts, hb, err = s.Connect(ctx, p, ""); err != nil {
		logger.Error("[authTCP.Connect() 失败] ", key, err)
		return
	}
	p.Op = protocol.OpAuthReply
	p.Body = nil
	if err = p.WriteTCP(wr); err != nil {
		logger.Error("authTCP.WriteTCP(key:%v).err(%v)", key, err)
		return
	}
	err = wr.Flush()
	return
}
