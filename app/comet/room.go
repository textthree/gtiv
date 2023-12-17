package comet

import (
	"sync"

	"gtiv/kit/impkg/protocol"
)

// Room is a room and store channel room info.
type Room struct {
	ID        string
	rLock     sync.RWMutex
	next      *Channel // 房间里面通过双链表来维护在线人数，发送群消息就会遍历这个链表中在线的 comet 进行发送
	drop      bool
	Online    int32 // dirty read is ok
	AllOnline int32
}

// NewRoom new a room struct, store channel room info.
func NewRoom(id string) (r *Room) {
	r = new(Room)
	r.ID = id
	r.drop = false
	r.next = nil
	r.Online = 0
	return
}

// Put put channel into the room. 新用户上线
func (r *Room) Put(ch *Channel) (err error) {
	r.rLock.Lock()
	if !r.drop {
		if r.next != nil {
			r.next.Prev = ch
		}
		ch.Next = r.next
		ch.Prev = nil
		r.next = ch // insert to header
		r.Online++
	} else {
		err = ErrRoomDroped
	}
	r.rLock.Unlock()
	return
}

// Del delete channel from the room. 用户下线
func (r *Room) Del(ch *Channel) bool {
	r.rLock.Lock()
	if ch.Next != nil {
		// if not footer
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil {
		// if not header
		ch.Prev.Next = ch.Next
	} else {
		r.next = ch.Next
	}
	ch.Next = nil
	ch.Prev = nil
	r.Online--
	r.drop = r.Online == 0
	r.rLock.Unlock()
	return r.drop
}

// Push push msg to the room, if chan full discard it.
// 遍历发送给房间中的人，如果房间人数太多发送会很慢，所以房间中的用户打散成多个 bucket 粒度，
// bucket.go 每个 bucket 下又有多个 goroutine（多个 worker）所以 goim 发送效率很高
// 但是每个群的人数还是不要太多
func (r *Room) Push(p *protocol.Proto) {
	r.rLock.RLock()
	for ch := r.next; ch != nil; ch = ch.Next {
		_ = ch.Push(p)
	}
	r.rLock.RUnlock()
}

// Close close the room.
func (r *Room) Close() {
	r.rLock.RLock()
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Close()
	}
	r.rLock.RUnlock()
}

// OnlineNum the room all online.
func (r *Room) OnlineNum() int32 {
	if r.AllOnline > 0 {
		return r.AllOnline
	}
	return r.Online
}
