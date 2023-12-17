package comet

import (
	"context"
	log "github.com/golang/glog"
	"github.com/text3cn/goodle/providers/etcd"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	logic "gtiv/app/imbiz/rpcapi"
	"gtiv/kit/impkg/logger"
	"math/rand"
	"os"
	"time"

	"github.com/zhenjl/cityhash"
	"google.golang.org/grpc"
)

const (
	minServerHeartbeat = time.Minute * 10
	maxServerHeartbeat = time.Minute * 30
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
	grpcMaxSendMsgSize        = 1 << 24
	grpcMaxCallMsgSize        = 1 << 24
	grpcKeepAliveTime         = time.Second * 10
	grpcKeepAliveTimeout      = time.Second * 3
)

var balancerCounter int = 0

// Server is comet server.
type Server struct {
	c          *Config
	round      *Round    // accept round store
	buckets    []*Bucket // subkey bucket
	bucketIdx  uint32
	serverID   string
	rpcClient  logic.LogicClient
	rpcClients []logic.LogicClient
}

//	protoc -I=. --go_out=. hello.proto
//
// NewServer returns a new Server.
func NewServer(cfg *Config) *Server {

	host, _ := os.Hostname()
	s := &Server{
		c:        cfg,
		round:    NewRound(cfg),
		serverID: host,
	}
	s.discoverLogicClient(cfg)
	go s.rpcClientLoadBalancer()
	// init bucket
	s.buckets = make([]*Bucket, cfg.Bucket.Size)
	s.bucketIdx = uint32(cfg.Bucket.Size)
	for i := 0; i < cfg.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(cfg.Bucket)
	}
	go s.onlineproc()
	return s
}

// 发现 logic 节点
func (self *Server) discoverLogicClient(cfg *Config) {
	var clients []logic.LogicClient
	etcd.Instance().ServiceDiscovery(cfg.LogicServiceName, func(cometNodes []string) {
		for _, rpcAddr := range cometNodes {
			// 每 10 秒换一次节点
			grpcClient, err := grpc.Dial(rpcAddr, []grpc.DialOption{
				grpc.WithInitialWindowSize(grpcInitialWindowSize),
				grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
				grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:                grpcKeepAliveTime,
					Timeout:             grpcKeepAliveTimeout,
					PermitWithoutStream: true,
				}),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			}...)
			if err != nil {
				panic(err)
			}
			client := logic.NewLogicClient(grpcClient)
			clients = append(clients, client)
		}
		// 不用加写锁，配置了负载均衡器的间隔事件与之差开
		self.rpcClients = clients
	})
}

// logic rpc 客户端负载均衡器
// 轮流取一个节点用，每 30 秒换一下
func (self *Server) rpcClientLoadBalancer() {
	nodeNum := len(self.rpcClients)
	if nodeNum > 0 {
		self.rpcClient = self.rpcClients[0]
	}
	balancerCounter = 0
	ticker := time.Tick(30 * time.Second)
	for {
		select {
		case <-ticker:
			balancerCounter++
			if nodeNum > 0 {
				index := balancerCounter % len(self.rpcClients)
				if balancerCounter == 9 {
					balancerCounter = 0
				}
				self.rpcClient = self.rpcClients[index]
			}
		}
	}
}

// Buckets return all buckets.
func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

// Bucket get the bucket by subkey.
func (s *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketIdx
	if Conf.Debug {
		logger.Infof("%s hit channel bucket index: %d use cityhash", subKey, idx)
	}
	return s.buckets[idx]
}

// RandServerHearbeat rand server heartbeat.
func (s *Server) RandServerHearbeat() time.Duration {
	return (minServerHeartbeat + time.Duration(rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat))))
}

// Close close the server.
func (s *Server) Close() (err error) {
	return
}

func (s *Server) onlineproc() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("onlineproc panic: ", r)
		}
	}()
	for {
		var (
			allRoomsCount map[string]int32
			err           error
		)
		roomCount := make(map[string]int32)
		for _, bucket := range s.buckets {
			for roomID, count := range bucket.RoomsCount() {
				roomCount[roomID] += count
			}
		}
		if allRoomsCount, err = s.RenewOnline(context.Background(), s.serverID, roomCount); err != nil {
			time.Sleep(time.Second)
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpRoomsCount(allRoomsCount)
		}
		time.Sleep(time.Second * 10)
	}
}
