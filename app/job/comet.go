package job

import (
	"context"
	"fmt"
	"github.com/text3cn/goodle/providers/etcd"
	"github.com/text3cn/goodle/providers/goodlog"
	"google.golang.org/grpc/credentials/insecure"
	cometpb "gtiv/app/comet/grpcapi"
	"os"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
)

var (
	// grpc options
	grpcKeepAliveTime    = time.Second * 10
	grpcKeepAliveTimeout = time.Second * 3
	grpcBackoffMaxDelay  = time.Second * 3
	grpcMaxSendMsgSize   = 1 << 24
	grpcMaxCallMsgSize   = 1 << 24
)

const (
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
)

// Comet is a cometpb.
type Comet struct {
	serverID      string
	client        cometpb.CometClient
	pushChan      []chan *cometpb.PushMsgReq
	roomChan      []chan *cometpb.BroadcastRoomReq
	broadcastChan chan *cometpb.BroadcastReq
	pushChanNum   uint64
	roomChanNum   uint64
	routineSize   uint64
	ctx           context.Context
	cancel        context.CancelFunc
}

func initComet(cfg *CometConf) map[string]*Comet {
	comets := map[string]*Comet{}
	hostname, _ := os.Hostname()
	var err error

	etcd.Instance().ServiceDiscovery(cfg.ServiceName, func(cometNodes []string) {
		for _, rpcAddr := range cometNodes {
			comet := &Comet{
				serverID:      hostname,
				pushChan:      make([]chan *cometpb.PushMsgReq, cfg.RoutineSize),
				roomChan:      make([]chan *cometpb.BroadcastRoomReq, cfg.RoutineSize),
				broadcastChan: make(chan *cometpb.BroadcastReq, cfg.RoutineSize),
				routineSize:   uint64(cfg.RoutineSize),
			}
			if comet.client, err = newCometClient(rpcAddr); err != nil {
				panic(err)
			}
			for i := 0; i < cfg.RoutineSize; i++ {
				comet.pushChan[i] = make(chan *cometpb.PushMsgReq, cfg.RoutineChan)
				comet.roomChan[i] = make(chan *cometpb.BroadcastRoomReq, cfg.RoutineChan)
				go comet.process(comet, i)
			}
			comets[hostname] = comet
		}
	})
	return comets
}

func newCometClient(addr string) (cometpb.CometClient, error) {
	// 建立 RPC 客户端
	grpcClient, err := grpc.Dial(addr, []grpc.DialOption{
		//grpc.WithInitialWindowSize(grpcInitialWindowSize),
		//grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
		//grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
		//grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
		//grpc.WithKeepaliveParams(keepalive.ClientParameters{
		//	Time:                grpcKeepAliveTime,
		//	Timeout:             grpcKeepAliveTimeout,
		//	PermitWithoutStream: true,
		//}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}...)
	if err != nil {
		return nil, err
	}
	return cometpb.NewCometClient(grpcClient), err
}

// Push push a user message.
func (c *Comet) Push(arg *cometpb.PushMsgReq) (err error) {
	idx := atomic.AddUint64(&c.pushChanNum, 1) % c.routineSize
	c.pushChan[idx] <- arg
	return
}

// BroadcastRoom broadcast a room message.
func (c *Comet) BroadcastRoom(arg *cometpb.BroadcastRoomReq) (err error) {
	idx := atomic.AddUint64(&c.roomChanNum, 1) % c.routineSize
	c.roomChan[idx] <- arg
	return
}

// Broadcast broadcast a message.
func (c *Comet) Broadcast(arg *cometpb.BroadcastReq) (err error) {
	c.broadcastChan <- arg
	return
}

func (c *Comet) process(comet *Comet, i int) {
	var pushChan chan *cometpb.PushMsgReq = comet.pushChan[i]
	var roomChan chan *cometpb.BroadcastRoomReq = comet.roomChan[i]
	var broadcastChan chan *cometpb.BroadcastReq = comet.broadcastChan
	comet.ctx, comet.cancel = context.WithCancel(context.Background())
	for {
		select {
		case broadcastArg := <-broadcastChan:
			_, err := c.client.Broadcast(context.Background(), &cometpb.BroadcastReq{
				Proto:   broadcastArg.Proto,
				ProtoOp: broadcastArg.ProtoOp,
				Speed:   broadcastArg.Speed,
			})
			if err != nil {
				err = fmt.Errorf("c.client.Broadcast(%s, reply) serverId:%s error(%v)", broadcastArg, c.serverID, err)
				goodlog.Greenf("出错了1 err=%v", err)
			}
		case roomArg := <-roomChan:
			_, err := c.client.BroadcastRoom(context.Background(), &cometpb.BroadcastRoomReq{
				RoomID: roomArg.RoomID,
				Proto:  roomArg.Proto,
			})
			if err != nil {
				err = fmt.Errorf("c.client.BroadcastRoom(%s, reply) serverId:%s error(%v)", roomArg, c.serverID, err)
				goodlog.Pink(err)
			}
		case pushArg := <-pushChan:
			_, err := c.client.PushMsg(context.Background(), &cometpb.PushMsgReq{
				Keys:    pushArg.Keys,
				Proto:   pushArg.Proto,
				ProtoOp: pushArg.ProtoOp,
			})
			if err != nil {
				err = fmt.Errorf("c.client.PushMsg(%s, reply) serverId:%s error(%v)", pushArg, c.serverID, err)
				goodlog.Pink(err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// Close close the resources.
func (c *Comet) Close() (err error) {
	finish := make(chan bool)
	go func() {
		for {
			n := len(c.broadcastChan)
			for _, ch := range c.pushChan {
				n += len(ch)
			}
			for _, ch := range c.roomChan {
				n += len(ch)
			}
			if n == 0 {
				finish <- true
				return
			}
			time.Sleep(time.Second)
		}
	}()
	select {
	case <-finish:
		goodlog.Green("close comet finish")
	case <-time.After(5 * time.Second):
		err = fmt.Errorf("close comet(server:%s push:%d room:%d broadcast:%d) timeout", c.serverID, len(c.pushChan), len(c.roomChan), len(c.broadcastChan))
	}
	c.cancel()
	return
}
