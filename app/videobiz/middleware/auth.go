package middleware

import (
	"errors"
	"github.com/text3cn/goodle/kit/cryptokit"
	"github.com/text3cn/goodle/providers/httpserver"
	"gtiv/app/videobiz/internal/boot"
)

func Auth() httpserver.MiddlewareHandler {
	return func(context *httpserver.Context) error {
		token, _ := context.Req.Header("Authorization")
		key := boot.TokenKey
		userId := cryptokit.DynamicDecrypt(key, token)
		if userId != "" {
			context.SetVal("userId", userId)
			context.Next()
			return nil
		}
		return errors.New("Videobiz Authorization failed.")
	}
}
