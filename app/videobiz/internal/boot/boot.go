package boot

import (
	"github.com/spf13/viper"
	"github.com/text3cn/goodle/providers/config"
)

var VideobizCfg *viper.Viper
var AppCfg *viper.Viper
var TokenKey string

func init() {
	// videobiz
	var err error
	VideobizCfg, err = config.Instance().LoadConfig("videobiz.yaml")
	if err != nil {
		panic(err)
	}
	TokenKey = VideobizCfg.GetString("tokenKey")

	// app
	AppCfg, err = config.Instance().LoadConfig("app.yaml")
	if err != nil {
		panic(err)
	}
}
