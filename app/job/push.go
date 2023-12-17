package job

import (
	"context"
	"fmt"
	"github.com/text3cn/goodle/providers/goodlog"
	cometpb "gtiv/app/comet/grpcapi"
	logicpb "gtiv/app/imbiz/rpcapi"
	"gtiv/kit/impkg/bytes"
	"gtiv/kit/impkg/protocol"
)

func (job *Job) push(ctx context.Context, pushMsg *logicpb.PushMsg) (err error) {
	switch pushMsg.Type {
	case logicpb.PushMsg_PUSH:
		//logger.Cgreen("job 处理个人消息 ")
		//logger.Ccyan(pushMsg)
		err = job.pushKeys(pushMsg.Operation, pushMsg.Server, pushMsg.Keys, pushMsg.Msg)
	case logicpb.PushMsg_ROOM:
		//logger.Cgreen("job 处理群消息 ")
		//logger.Ccyan(pushMsg)
		err = job.getRoom(pushMsg.Room).Push(pushMsg.Operation, pushMsg.Msg)
	case logicpb.PushMsg_BROADCAST:
		err = job.broadcast(pushMsg.Operation, pushMsg.Msg, pushMsg.Speed)
	default:
		err = fmt.Errorf("no match push type: %s", pushMsg.Type)
	}
	if err != nil {
		goodlog.Error(err.Error())
	}
	return
}

// pushKeys push a message to a batch of subkeys.
func (j *Job) pushKeys(operation int32, serverID string, subKeys []string, body []byte) (err error) {
	buf := bytes.NewWriterSize(len(body) + 64)
	p := &protocol.Proto{
		Ver:  1,
		Op:   operation,
		Body: body,
	}

	p.WriteTo(buf)
	p.Body = buf.Buffer()
	p.Op = protocol.OpRaw
	var args = cometpb.PushMsgReq{
		Keys:    subKeys,
		ProtoOp: operation,
		Proto:   p,
	}
	// 直接从 redis 找到这个用户的连接在哪个 comet 上然后发送过去
	if c, ok := j.cometServers[serverID]; ok {
		if err = c.Push(&args); err != nil {
			goodlog.Errorf("c.Push(%v) serverID:%s error(%v)", args, serverID, err)
		}
		goodlog.Infof("pushKey:%s comets:%d", serverID, len(j.cometServers))
	} else {
		goodlog.Error("Unable to find comet node")
	}
	return
}

// broadcast broadcast a message to all.
func (j *Job) broadcast(operation int32, body []byte, speed int32) (err error) {
	buf := bytes.NewWriterSize(len(body) + 64)
	p := &protocol.Proto{
		Ver:  1,
		Op:   operation,
		Body: body,
	}
	p.WriteTo(buf)
	p.Body = buf.Buffer()
	p.Op = protocol.OpRaw
	comets := j.cometServers
	speed /= int32(len(comets))
	var args = cometpb.BroadcastReq{
		ProtoOp: operation,
		Proto:   p,
		Speed:   speed,
	}
	for serverID, c := range comets {
		if err = c.Broadcast(&args); err != nil {
			goodlog.Errorf("c.Broadcast(%v) serverID:%s error(%v)", args, serverID, err)
		}
	}
	goodlog.Infof("broadcast comets:%d", len(comets))
	return
}

// broadcastRoomRawBytes broadcast aggregation messages to room.
func (j *Job) broadcastRoomRawBytes(roomID string, body []byte) (err error) {
	args := cometpb.BroadcastRoomReq{
		RoomID: roomID,
		Proto: &protocol.Proto{
			Ver:  1,
			Op:   protocol.OpRaw,
			Body: body,
		},
	}
	comets := j.cometServers
	for serverID, c := range comets {
		if err = c.BroadcastRoom(&args); err != nil {
			goodlog.Errorf("c.BroadcastRoom(%v) roomID:%s serverID:%s error(%v)", args, roomID, serverID, err)
		}
	}
	goodlog.Infof("broadcastRoom comets:%d", len(comets))
	return
}
