package comet

import (
	"gtiv/kit/impkg/bufio"
	"gtiv/kit/impkg/logger"
	"gtiv/kit/impkg/protocol"
	"sync"
)

// 消息读写的通道，comet 的内部是通过管道来实现的，job 将消息塞进到 signal
// Channel 双向链表结构持有长连接
type Channel struct {
	Room     *Room
	CliProto Ring                 // cliProto 是一个 Ring Buffer，保存 Room 广播或是直接发送过来的消息体。
	signal   chan *protocol.Proto // 接收来自 job 的 protobuffer ？
	Writer   bufio.Writer
	Reader   bufio.Reader
	Next     *Channel
	Prev     *Channel
	Mid      int64
	Key      string
	IP       string
	watchOps map[int32]struct{} // 接收哪些 opration
	mutex    sync.RWMutex
}

// 启动 TCP 或 WS 服务时创建通道
func NewChannel(cli, svr int) *Channel {
	c := new(Channel)
	c.CliProto.Init(cli)
	c.signal = make(chan *protocol.Proto, svr)
	c.watchOps = make(map[int32]struct{})
	return c
}

// Watch watch a operation.
// 现在是把群 id 当成 operation 来用了，operation 是全局大广播，针对所有用户
func (this *Channel) Watch(accepts ...int32) {
	this.mutex.Lock()
	for _, operation := range accepts {
		this.AddWatch(operation)
	}
	this.mutex.Unlock()
}

// 目前 operation 就是群 id
func (this *Channel) AddWatch(operation int32) {
	this.watchOps[operation] = struct{}{}
}

// UnWatch unwatch an operation
func (c *Channel) UnWatch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		delete(c.watchOps, op)
	}
	c.mutex.Unlock()
}

// NeedPush verify if in watch.
func (c *Channel) NeedPush(op int32) bool {
	c.mutex.RLock()
	if _, ok := c.watchOps[op]; ok {
		c.mutex.RUnlock()
		return true
	}
	c.mutex.RUnlock()
	return false
}

// 将 job 发来的消息塞进管道
func (c *Channel) Push(p *protocol.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
		err = ErrSignalFullMsgDropped
		logger.Cred(ErrSignalFullMsgDropped)
	}
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() *protocol.Proto {
	return <-c.signal
}

// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {

	c.signal <- protocol.ProtoReady
}

// Close close the channel.
func (c *Channel) Close() {
	c.signal <- protocol.ProtoFinish
}
