package main

import (
	"fmt"
	"github.com/text3cn/goodle/core"
	"github.com/text3cn/goodle/goodle"
	"github.com/text3cn/goodle/providers/goodlog"
	"github.com/text3cn/goodle/providers/orm"
	"gtiv/app/videobiz/entity"
	"gtiv/app/videobiz/internal"
	"os/exec"
)

// @title Goodle API
// @description Goodle API Swagger 2.0
// @BasePath /
func main() {
	goodle.Init(func() {
		//autoMigrate(c)
		genapi()
	}).RunHttp(internal.Router)
}

func autoMigrate(c core.Container) {
	db := orm.GetDB()
	err := db.AutoMigrate(
		&entity.Video{},
		&entity.LiveRoom{},
		&entity.VideoSupport{},
		&entity.VideoCollect{},
		&entity.Follow{},
	)
	if err == nil {
		goodlog.Pink("migrate ok")
	}
}

func genapi() {
	command := exec.Command("/Users/t3/go/bin/swag", "init")
	output, err := command.Output()
	if err != nil {
		goodlog.Pinkf("swag init error: %v", err.Error())
	}
	fmt.Println(string(output))
}
