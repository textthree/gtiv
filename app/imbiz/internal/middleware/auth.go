package middleware

import (
	"errors"
	"github.com/text3cn/goodle/kit/cryptokit"
	"github.com/text3cn/goodle/providers/httpserver"
	"gtiv/app/imbiz/boot"
)

func Auth() httpserver.MiddlewareHandler {
	return func(context *httpserver.Context) error {
		language, _ := context.Req.Header("LanguageCode")
		context.SetVal("lng", language)
		lastLoginTime, _ := context.Req.Header("LastLoginTime")
		context.SetVal("lastLoginTime", lastLoginTime)
		passRoutes := []string{
			"/get-comet",
			"/user/login",
			"/user/register",
		}
		for _, element := range passRoutes {
			if element == context.Req.Uri() {
				context.Next()
				return nil
			}
		}
		token, _ := context.Req.Header("Authorization")
		userId := cryptokit.DynamicDecrypt(boot.TokenKey, token)
		if userId != "" && userId != "Token decryption invalid" {
			context.SetVal("uid", userId)
			context.Next()
			return nil
		}
		return errors.New("Goodle Authorization failed.")
	}
}
