package webrtc

import (
	"github.com/spf13/viper"
	"github.com/text3cn/goodle/kit/filekit"
	"os"
	"path/filepath"
)

// turn
var (
	turnListenIp string
	turnPublicIp string
	turnPort     string
	turnUser     string
	turnPwd      string
	turnRealm    string // 领域，把这个当做数据库名好了
)

var (
	signalListenAddress string
)

var (
	redisAddr string
	redisPwd  string
	redisDb   int
)

func InitConfig() {
	configFileName := "config.toml"
	currentPath, _ := os.Getwd()
	if exists, _ := filekit.PathExists(filepath.Join(currentPath, "config_local.toml")); exists {
		configFileName = "config_local.toml"
	}
	cfg := viper.New()
	cfg.AddConfigPath("./")
	cfg.SetConfigName(configFileName)
	cfg.SetConfigType("toml")
	if err := cfg.ReadInConfig(); err != nil {
		panic(err)
	}
	turnListenIp = cfg.GetString("turn.listenIp")
	turnPublicIp = cfg.GetString("turn.publicIp")
	turnPort = cfg.GetString("turn.port")
	turnUser = cfg.GetString("turn.user")
	turnPwd = cfg.GetString("turn.pwd")
	turnRealm = cfg.GetString("turn.realm")
	signalListenAddress = cfg.GetString("signal.listenAddress")
	redisAddr = cfg.GetString("redis.addr")
	redisPwd = cfg.GetString("redis.password")
	redisDb = cfg.GetInt("redis.db")
}
