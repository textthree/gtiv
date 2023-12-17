package im

import (
	"context"
	"gtiv/app/imbiz/im/conf"
	"net"

	"time"

	pb "gtiv/app/imbiz/rpcapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	// use gzip decoder
	_ "google.golang.org/grpc/encoding/gzip"
)

// New im grpc server
func NewGrpcServer(cfg *conf.RPCServer, l *Logic) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(cfg.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(cfg.ForceCloseWait),
		Time:                  time.Duration(cfg.KeepAliveInterval),
		Timeout:               time.Duration(cfg.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(cfg.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)
	pb.RegisterLogicServer(srv, &server{srv: l})
	lis, err := net.Listen(cfg.Network, cfg.Addr)
	if err != nil {
		panic(err)
	}
	go func() {
		if err = srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return srv
}

type server struct {
	srv *Logic
	*pb.UnimplementedLogicServer
}

// Connect connect a conn.
func (s *server) Connect(ctx context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {
	mid, key, room, accepts, hb, err := s.srv.Connect(ctx, req.Server, req.Cookie, req.Token)
	if err != nil {
		return &pb.ConnectReply{}, err
	}
	return &pb.ConnectReply{Mid: mid, Key: key, RoomID: room, Accepts: accepts, Heartbeat: hb}, nil
}

// Disconnect disconnect a conn.
func (s *server) Disconnect(ctx context.Context, req *pb.DisconnectReq) (*pb.DisconnectReply, error) {
	has, err := s.srv.Disconnect(ctx, req.Mid, req.Key, req.Server)
	if err != nil {
		return &pb.DisconnectReply{}, err
	}
	return &pb.DisconnectReply{Has: has}, nil
}

// Heartbeat beartbeat a conn.
func (s *server) Heartbeat(ctx context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	if err := s.srv.Heartbeat(ctx, req.Mid, req.Key, req.Server); err != nil {
		return &pb.HeartbeatReply{}, err
	}
	return &pb.HeartbeatReply{}, nil
}

// RenewOnline renew server online.
func (s *server) RenewOnline(ctx context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	allRoomCount, err := s.srv.RenewOnline(ctx, req.Server, req.RoomCount)
	if err != nil {
		return &pb.OnlineReply{}, err
	}
	return &pb.OnlineReply{AllRoomCount: allRoomCount}, nil
}
