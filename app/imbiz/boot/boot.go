package boot

import (
	"github.com/spf13/viper"
	"github.com/text3cn/goodle/providers/config"
)

var ImbizCfg *viper.Viper
var AppCfg *viper.Viper
var TokenKey string

func init() {
	// imbiz
	var err error
	ImbizCfg, err = config.Instance().LoadConfig("imbiz.yaml")
	if err != nil {
		panic(err)
	}
	TokenKey = ImbizCfg.GetString("tokenKey")

	// app
	AppCfg, err = config.Instance().LoadConfig("app.yaml")
	if err != nil {
		panic(err)
	}
}
