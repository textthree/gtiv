package job

import (
	"context"
	"encoding/json"
	cluster "github.com/bsm/sarama-cluster" // kafka 客户端
	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/text3cn/goodle/kit/strkit"
	"github.com/text3cn/goodle/providers/goodlog"
	pb "gtiv/app/imbiz/rpcapi"
	"sync"
	"time"
)

// Job is push job.
type Job struct {
	config       *Config
	consumer     *cluster.Consumer
	cometServers map[string]*Comet
	redis        *redis.Pool
	// mysql.Connect(Mysql_log.DbName, Mysql_log)
	rooms      map[string]*Room
	roomsMutex sync.RWMutex
}

// New new a push job.
func New(config *Config) *Job {
	j := &Job{
		config:   config,
		rooms:    make(map[string]*Room),
		consumer: newKafkaSub(config.Kafka),
		redis:    newRedis(config.Redis),
		// 原先 comet 是通过 discovery 发现的：https://github.com/Terry-Mao/goim/blob/master/internal/job/job.go
		cometServers: initComet(config.Comet),
	}
	return j
}

func newKafkaSub(c *Kafka) *cluster.Consumer {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	consumer, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Topic}, config)
	if err != nil {
		panic(err)
	}
	return consumer
}

func newRedis(c *Redis) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.Idle,
		MaxActive:   c.Active,
		IdleTimeout: time.Duration(c.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(c.Network, c.Addr,
				redis.DialConnectTimeout(time.Duration(c.DialTimeout)),
				redis.DialReadTimeout(time.Duration(c.ReadTimeout)),
				redis.DialWriteTimeout(time.Duration(c.WriteTimeout)),
				redis.DialPassword(c.Auth),
			)
			if err != nil {
				return nil, err
			}
			if c.SelectDb > 0 {
				if _, err = conn.Do("SELECT", c.SelectDb); err != nil {
					conn.Close()
				}
			}
			return conn, nil
		},
	}
}

// Close close resounces.
func (j *Job) Close() error {
	if j.consumer != nil {
		return j.consumer.Close()
	}
	return nil
}

// Consume messages, watch signals
func (self *Job) Consume() {
	for {
		select {
		case err := <-self.consumer.Errors():
			goodlog.Errorf("consumer error(%v)", err)
			// 如果报 kafka 连接到本地 ip，可能是 kafka 集群配置没有对外公开 ip
		case n := <-self.consumer.Notifications():
			goodlog.Info("[consumer rebalanced]", n)
		case msg, ok := <-self.consumer.Messages():
			if !ok {
				goodlog.Error("<-self.consumer.Messages() not OK")
				return
			}
			self.consumer.MarkOffset(msg, "")
			// 往 comet 推消息
			pushMsg := new(pb.PushMsg)
			if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
				log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
				continue
			}
			//logger.Green("pushMsg 消费", pushMsg)
			if err := self.push(context.Background(), pushMsg); err != nil {
				goodlog.Error("j.push() error", pushMsg, err)
			}
			handleMessage(pushMsg, self)
		}
	}
}

func handleMessage(pushMsg *pb.PushMsg, job *Job) {
	//fmt.Println(msg.Topic, msg.Partition, msg.Offset, msg.Key, pushMsg)
	//logger.Cpink(pushMsg.Operation)
	//logger.Cpink(pushMsg.Msg)
	/*go fn.SafeGo(func() {
		sql := `INSERT INTO message_2207 (room_id, from_user, message_type, message)
			VALUES('2', '4', '1', '` + msgStruct.Content + `')`
		mysql.GetRawDb().Exec(sql)
	})*/

	// 消息体序列化
	var msgMap map[string]interface{}
	if err := json.Unmarshal([]byte(pushMsg.Msg), &msgMap); err != nil {
		goodlog.Error("json unmarshal error")
	}
	// 撤回消息
	msgType, _ := msgMap["MsgType"].(float64)
	if msgType == 6 {
		conn := job.redis.Get()
		//defer conn.Close()
		replyMsgId, _ := msgMap["Content"].(string)
		roomId, _ := msgMap["RoomId"].(string)
		key, _ := RoomMessage("room_" + roomId)
		rows := 50
		for page := 1; page <= 100; page++ {
			start := (page-1)*rows + 1
			end := start + rows - 1
			// 从倒数第 start 条取到倒数第 end 条
			r, _ := conn.Do("LRANGE", key, -end, -start)
			res, _ := r.([]interface{})
			for _, v := range res {
				it, _ := v.([]uint8)
				if err := json.Unmarshal(it, &msgMap); err != nil {
					goodlog.Error("repeal message body unmarshal error")
				}
				msgId, _ := msgMap["MsgId"].(string)
				if msgId == replyMsgId {
					goodlog.Error(replyMsgId)
					// 如果消息在后 50 页则，从后往前搜索删除，反之从头往尾找
					if page <= 50 {
						conn.Do("LREM", key, -1, v)
					} else {
						conn.Do("LREM", key, 1, v)
					}
					break
				}
			}
		}
	}
}

// 此方法从 im/internal/cache/rediskey/rediskey.go 中复制而来
// 群消息，失效策略是最多保存 5000 条，如果一个群半年不活跃则这 5000 条数据失效
func RoomMessage(roomId string) (string, int) {
	if !strkit.StartWith(roomId, "room_") {
		// 增加 room_ 前缀，主要是兼容 comet 连接的 room bucket key
		roomId = "room_" + roomId
	}
	return "room:message:" + roomId, 3600 * 24 * 180
}

func test() {}
