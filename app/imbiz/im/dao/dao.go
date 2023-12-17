package dao

import (
	"context"
	"github.com/text3cn/goodle/providers/redis"
	kafka "gopkg.in/Shopify/sarama.v1"
	"gtiv/app/imbiz/im/conf"
)

// Dao dao.
type Dao struct {
	c           *conf.Config
	kafkaPub    kafka.SyncProducer
	redis       *redis.Client
	redisExpire int32
}

// New new a dao and return.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:           c,
		kafkaPub:    newKafkaPub(c.Kafka),
		redis:       newRedis(),
		redisExpire: 3600 * 24,
	}
	return d
}

func newKafkaPub(c *conf.Kafka) kafka.SyncProducer {
	kc := kafka.NewConfig()
	kc.Producer.RequiredAcks = kafka.WaitForAll // Wait for all in-sync replicas to ack the message
	kc.Producer.Retry.Max = 10                  // Retry up to 10 times to produce the message
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return pub
}

func newRedis() *redis.Client {
	return redis.Instance().Conn()
}

// Close close the resource.
func (d *Dao) Close() error {
	return d.redis.Close()
}

// Ping dao ping.
func (d *Dao) Ping(c context.Context) error {
	return d.pingRedis(c)
}
