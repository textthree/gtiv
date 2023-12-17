### Bucket
每个 Comet 程序拥有若干个 Bucket, 可以理解为 Session Management, 保存着当前 Comet 服务于哪些 Room 和 Channel.
长连接具体分布在哪个 Bucket 上，根据 SubKey 一致性 Hash 来选择。

Bucket 结构体维护当前消息通道和房间的信息，提供方法也很简单，就是加、减、统计 Channel 和 Room，
一个 Comet Server 默认开启 1024 Bucket, 这样做的好处是减少锁 ( Bucket.cLock ) 争用，在大并发业务上尤其明显。

bucket 按照 key 的 cityhash 求余命中的，没有用一致性hash，因为这里不涉及迁移

### Room
每个 Room 内维护 N 个 Channel，在该 Room 内广播消息，会发送给房间内的所有 Channel.

Room 结构体不但要维护所属的消息通道 Channel, 还要消息广播的合并写，即 Batch Write, 如果不合并写，
每来一个小的消息都通过长连接写出去，系统 Syscall 调用的开销会非常大，Pprof 的时候会看到网络 Syscall 是大头，
Logic Server 通过 RPC 调用将广播的消息发给 Room.Push，数据会被暂存在 proto 这个结构体里，
每个 Room 在初始化时会开启一个 groutine 用来处理暂存的消息，达到 Batch Num 数量或是延迟一定时间后，将消息批量 Push 到 Channel 消息通道，
代价就是消息有一点秒的延迟，具体延迟多久在代码中通过多维度计算出来的，1 秒左右，当然你也可以改造写死 500 毫秒一次合并，
最后将消息写到 Channel Ring Buffer

### Channel
每个 Channel 维护一个长连接用户，只能对应一个 Room，推送的消息可以在 Room 内广播，也可以推送到指定的 Channel.
就是对网络 Writer/Reader 的封装，cliProto 是一个 Ring Buffer，保存 Room 广播或是直接发送过来的消息体。
新连接上线时 Channel 开启一个 dispatchTCP groutine, 阻塞等待 Ring Buffer 数据。


### Proto
消息结构体，存放版本号，操作类型，消息序号和消息体。


### 补充 - 来自毛剑
- 私信发送其实也有合并的，和 room 合并不同的是，在 ringbuffer 取消息饥饿时候才会真正 flush

- 还有一个优化可以改进，因为 room 有个特点大家消息可能都一样，所以在 room 提前合并成字节 buffer，然后广播所有人，
避免每个人都序列化一次，然后利用gc来处理这个 buffer 的释放，这样可以节省大量 cpu，目前这个优化还没做。