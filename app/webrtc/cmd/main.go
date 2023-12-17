package main

import (
	"github.com/text3cn/goodle/goodle"
	"gtiv/app/webrtc"
)

func main() {
	goodle.Init()
	webrtc.InitConfig()
	go webrtc.StartSingnalServer()
	webrtc.PionTurnServer()
}
