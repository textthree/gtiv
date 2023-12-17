package comet

import (
	"time"

	"github.com/BurntSushi/toml"
	xtime "gtiv/kit/impkg/time"
)

var (
	Conf *Config
)

func Init() (err error) {
	Conf = Default()
	_, err = toml.DecodeFile("./comet.toml", &Conf)
	return
}

// Default new a config with specified defualt value.
func Default() *Config {
	return &Config{
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              ":3109",
			Timeout:           xtime.Duration(time.Second),
			IdleTimeout:       xtime.Duration(time.Second * 60),
			MaxLifeTime:       xtime.Duration(time.Hour * 2),
			ForceCloseWait:    xtime.Duration(time.Second * 20),
			KeepAliveInterval: xtime.Duration(time.Second * 60),
			KeepAliveTimeout:  xtime.Duration(time.Second * 20),
		},
		TCP: &TCP{
			Bind:         []string{":3101"},
			Sndbuf:       4096,
			Rcvbuf:       4096,
			KeepAlive:    false,
			Reader:       32,
			ReadBuf:      1024,
			ReadBufSize:  8192,
			Writer:       32,
			WriteBuf:     1024,
			WriteBufSize: 8192,
		},
		Websocket: &Websocket{
			Bind: []string{":3102"},
		},
		Protocol: &Protocol{
			Timer:            32,
			TimerSize:        2048,
			CliProto:         5,
			SvrProto:         10,
			HandshakeTimeout: xtime.Duration(time.Second * 5),
		},
		Bucket: &BucketConf{
			Size:          32,
			Channel:       1024,
			Room:          1024,
			RoutineAmount: 32,
			RoutineSize:   1024,
		},
	}
}

// Config is comet config.
type Config struct {
	Debug            bool
	TCP              *TCP
	Websocket        *Websocket
	Protocol         *Protocol
	Bucket           *BucketConf
	RPCServer        *RPCServer
	Whitelist        *WhitelistConf
	LogicRpc         []string
	LogicServiceName string
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

// TCP is tcp config.
type TCP struct {
	Bind         []string
	Sndbuf       int
	Rcvbuf       int
	KeepAlive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// Websocket is websocket config.
type Websocket struct {
	Bind        []string
	TLSOpen     bool
	TLSBind     []string
	CertFile    string
	PrivateFile string
}

// Protocol is protocol config.
type Protocol struct {
	Timer            int
	TimerSize        int
	SvrProto         int
	CliProto         int
	HandshakeTimeout xtime.Duration
}

// Bucket is bucket config.
type BucketConf struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

// Whitelist is white list config.
type WhitelistConf struct {
	Whitelist []int64
	WhiteLog  string
}
