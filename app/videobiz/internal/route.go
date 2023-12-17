package internal

import (
	"github.com/text3cn/goodle/providers/httpserver"
	"gtiv/app/videobiz/internal/api"
	"gtiv/app/videobiz/internal/api/user"
	"gtiv/app/videobiz/internal/api/video"
	"gtiv/app/videobiz/middleware"
)

func Router(core *httpserver.Engine) {
	core.UseMiddleware(middleware.Auth())
	core.UseCross()
	// 静态路由
	core.Get("/", Index)
	core.Get("/live/list", api.LiveList)

	userGroup := core.Prefix("/user")
	{
		userGroup.Post("/register-by-username", user.RegisterByUsername)
		userGroup.Get("/video-master-info", user.VideoMasterInfo)
		userGroup.Get("/support-videos", user.SupportVideos)
		userGroup.Get("/creation-videos", user.CreationVideos)
		userGroup.Get("/collect-videos", user.CollectVideos)
		userGroup.Get("/follow", user.Follow)
		userGroup.Get("/follow-list", user.FollowList)
		userGroup.Get("/fans-list", user.FansList)
		userGroup.Get("/support-list", user.SupportList)
	}

	// video
	core.Get("/video/home", api.HomeVideo)
	core.Post("/video/support", video.Support)
	core.Post("/video/collect", video.Collect)
	core.Post("/video/incr-share-count", video.IncrShareNum)
	core.Get("/video/info", video.Info)

	// video upload
	core.Post("/video-upload", video.UploadVideo)
	core.Post("/video-reupload", video.ReUploadVideo)
}

func Index(ctx *httpserver.Context) {
	ctx.Resp.Json("videbiz index")
}
