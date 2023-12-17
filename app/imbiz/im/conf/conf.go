package conf

import (
	"gtiv/app/imbiz/boot"
	"time"

	xtime "gtiv/kit/impkg/time"
)

func InitConfig() *Config {
	serverTimeout := time.Duration(boot.AppCfg.GetInt("rpcServer.timeout"))
	clientTimeout := time.Duration(boot.AppCfg.GetInt("rpcClient.dial"))
	clientDial := time.Duration(boot.AppCfg.GetInt("rpcClient.timeout"))
	cfg := &Config{
		RPCClient: &RPCClient{
			Dial:    xtime.Duration(time.Second * clientTimeout),
			Timeout: xtime.Duration(time.Second * clientDial),
		},
		RPCServer: &RPCServer{
			Network:           boot.AppCfg.GetString("rpcServer.network"),
			Addr:              boot.AppCfg.GetString("rpcServer.addr"),
			Timeout:           xtime.Duration(time.Second * serverTimeout),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
		Backoff: &Backoff{MaxDelay: 300, BaseDelay: 3, Factor: 1.8, Jitter: 1.3},
		Kafka: &Kafka{
			Topic:   boot.AppCfg.GetString("kafka.topic"),
			Brokers: boot.AppCfg.GetStringSlice("kafka.brokers"),
		},
	}
	return cfg
}

// Config config.
type Config struct {
	RPCClient *RPCClient
	RPCServer *RPCServer
	Kafka     *Kafka
	Backoff   *Backoff
}

// Backoff backoff.
type Backoff struct {
	MaxDelay  int32
	BaseDelay int32
	Factor    float32
	Jitter    float32
}

// Kafka .
type Kafka struct {
	Topic   string
	Brokers []string
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    xtime.Duration
	Timeout xtime.Duration
}

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           xtime.Duration
	IdleTimeout       xtime.Duration
	MaxLifeTime       xtime.Duration
	ForceCloseWait    xtime.Duration
	KeepAliveInterval xtime.Duration
	KeepAliveTimeout  xtime.Duration
}
