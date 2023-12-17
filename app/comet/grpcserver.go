package comet

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gtiv/app/comet/grpcapi"
	"gtiv/kit/impkg/logger"
	"net"
	"time"
)

// New comet grpc server.
func New(c *RPCServer, s *Server) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)
	grpcapi.RegisterCometServer(srv, &server{s})
	lis, err := net.Listen(c.Network, c.Addr)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return srv
}

type server struct {
	srv *Server
}

var _ grpcapi.CometServer = &server{}

// PushMsg push a message to specified sub keys.
func (s *server) PushMsg(ctx context.Context, req *grpcapi.PushMsgReq) (reply *grpcapi.PushMsgReply, err error) {

	if len(req.Keys) == 0 || req.Proto == nil {
		return nil, ErrPushMsgArg
	}
	for _, key := range req.Keys {
		bucket := s.srv.Bucket(key)
		if bucket == nil {
			continue
		}
		if channel := bucket.Channel(key); channel != nil {
			if !channel.NeedPush(req.ProtoOp) {
				continue
			}
			if err = channel.Push(req.Proto); err != nil {
				logger.Ccyan("Push出错222")
				return
			}
		}
	}
	return &grpcapi.PushMsgReply{}, nil
}

// Broadcast broadcast msg to all user.
func (s *server) Broadcast(ctx context.Context, req *grpcapi.BroadcastReq) (*grpcapi.BroadcastReply, error) {
	if req.Proto == nil {
		return nil, ErrBroadCastArg
	}
	// TODO use broadcast queue
	go func() {
		for _, bucket := range s.srv.Buckets() {
			bucket.Broadcast(req.GetProto(), req.ProtoOp)
			if req.Speed > 0 {
				t := bucket.ChannelCount() / int(req.Speed)
				time.Sleep(time.Duration(t) * time.Second)
			}
		}
	}()
	return &grpcapi.BroadcastReply{}, nil
}

// BroadcastRoom broadcast msg to specified room.
func (s *server) BroadcastRoom(ctx context.Context, req *grpcapi.BroadcastRoomReq) (*grpcapi.BroadcastRoomReply, error) {
	if req.Proto == nil || req.RoomID == "" {
		return nil, ErrBroadCastRoomArg
	}
	for _, bucket := range s.srv.Buckets() {
		bucket.BroadcastRoom(req)
	}
	return &grpcapi.BroadcastRoomReply{}, nil
}

// Rooms gets all the room ids for the server.
func (s *server) Rooms(ctx context.Context, req *grpcapi.RoomsReq) (*grpcapi.RoomsReply, error) {
	var (
		roomIds = make(map[string]bool)
	)
	for _, bucket := range s.srv.Buckets() {
		for roomID := range bucket.Rooms() {
			roomIds[roomID] = true
		}
	}
	return &grpcapi.RoomsReply{Rooms: roomIds}, nil
}
