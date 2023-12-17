package job

import (
	"errors"
	"github.com/text3cn/goodle/providers/goodlog"
	"gtiv/kit/impkg/bytes"
	"gtiv/kit/impkg/protocol"
	"time"
)

var (
	// ErrComet commet error.
	ErrComet = errors.New("comet rpc is not available")
	// ErrCometFull comet chan full.
	ErrCometFull = errors.New("comet proto chan full")
	// ErrRoomFull room chan full.
	ErrRoomFull = errors.New("room proto chan full")

	roomReadyProto = new(protocol.Proto)
)

// Room room.
type Room struct {
	c     *RoomConf
	job   *Job
	id    string
	proto chan *protocol.Proto
}

// NewRoom new a room struct, store channel room info.
func NewRoom(job *Job, id string, c *RoomConf) (r *Room) {
	r = &Room{
		c:     c,
		id:    id,
		job:   job,
		proto: make(chan *protocol.Proto, c.Batch*2),
	}
	go r.pushproc(c.Batch, time.Duration(c.Signal))
	return
}

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(op int32, msg []byte) (err error) {
	var p = &protocol.Proto{
		Ver:  1,
		Op:   op,
		Body: msg,
	}
	select {
	case r.proto <- p:
	default:
		err = ErrRoomFull
	}
	return
}

// pushproc merge proto and push msgs in batch.
func (r *Room) pushproc(batch int, sigTime time.Duration) {
	var (
		n    int
		last time.Time
		p    *protocol.Proto
		buf  = bytes.NewWriterSize(int(protocol.MaxBodySize))
	)
	goodlog.Info("start room goroutine, roomId: ", r.id)
	td := time.AfterFunc(sigTime, func() {
		select {
		case r.proto <- roomReadyProto:
		default:
		}
	})
	defer td.Stop()
	for {
		if p = <-r.proto; p == nil {
			break // exit
		} else if p != roomReadyProto {
			// merge buffer ignore error, always nil
			p.WriteTo(buf)
			if n++; n == 1 {
				last = time.Now()
				td.Reset(sigTime)
				continue
			} else if n < batch {
				if sigTime > time.Since(last) {
					continue
				}
			}
		} else {
			if n == 0 {
				break
			}
		}
		_ = r.job.broadcastRoomRawBytes(r.id, buf.Buffer())
		// TODO use reset buffer
		// after push to room channel, renew a buffer, let old buffer gc
		buf = bytes.NewWriterSize(buf.Size())
		n = 0
		if r.c.Idle != 0 {
			td.Reset(time.Duration(r.c.Idle))
		} else {
			td.Reset(time.Minute)
		}
	}
	r.job.delRoom(r.id)
	goodlog.Info("room goroutine exit, roomId: ", r.id)
}

func (j *Job) delRoom(roomID string) {
	j.roomsMutex.Lock()
	delete(j.rooms, roomID)
	j.roomsMutex.Unlock()
}

func (j *Job) getRoom(roomID string) *Room {
	j.roomsMutex.RLock()
	room, ok := j.rooms[roomID]
	j.roomsMutex.RUnlock()
	// 房间不存在创建一个，并为其分配一个协程
	if !ok {
		j.roomsMutex.Lock()
		if room, ok = j.rooms[roomID]; !ok {
			room = NewRoom(j, roomID, j.config.Room)
			j.rooms[roomID] = room
		}
		j.roomsMutex.Unlock()
		goodlog.Info("new a room "+roomID+" active: ", len(j.rooms))
	}
	return room
}
