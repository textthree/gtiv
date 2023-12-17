package job

import (
	"github.com/text3cn/goodle/kit/filekit"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	xtime "gtiv/kit/impkg/time"
)

var (
	host string
	Conf *Config
)

func init() {
	//host, _ = os.Hostname()
}

// Init init config.
func Init() (err error) {
	Conf = Default()
	configFileName := "./job.toml"
	currentPath, _ := os.Getwd()
	if exists, _ := filekit.PathExists(filepath.Join(currentPath, "./job_alpha.toml")); exists {
		configFileName = "job_alpha.toml"
	}
	_, err = toml.DecodeFile(configFileName, &Conf)
	return
}

// Default new a config with specified defualt value.
func Default() *Config {
	return &Config{
		Comet: &CometConf{RoutineChan: 1024, RoutineSize: 32},
		Room: &RoomConf{
			Batch:  20,
			Signal: xtime.Duration(time.Second),
			Idle:   xtime.Duration(time.Minute * 15),
		},
	}
}

// Config is job config.
type Config struct {
	Kafka *Kafka
	Comet *CometConf
	Room  *RoomConf
	Redis *Redis
}

// Room is room config.
type RoomConf struct {
	Batch  int
	Signal xtime.Duration
	Idle   xtime.Duration
}

// Comet is comet config.
type CometConf struct {
	RoutineChan int
	RoutineSize int
	ServiceName string
}

// Kafka is kafka config.
type Kafka struct {
	Topic   string
	Group   string
	Brokers []string
}

// Redis is redis config.
type Redis struct {
	Network      string
	Addr         string
	Auth         string
	Active       int
	Idle         int
	DialTimeout  xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
	IdleTimeout  xtime.Duration
	Expire       xtime.Duration
	SelectDb     int
}
