package kit

import (
	"github.com/spf13/cast"
	"github.com/text3cn/goodle/providers/httpserver"
	"gtiv/types"
	"gtiv/types/errcode"
)

// 业务失败返回资源
// args[0] message
// args[1] code
func FailResponse(ctx *httpserver.Context, args ...interface{}) {
	msg := "fail"
	code := errcode.Fail
	var ok bool
	if len(args) > 0 {
		msg, ok = args[0].(string)
		if !ok {
			var err error
			err, ok = args[0].(error)
			if ok {
				msg = cast.ToString(err.Error())
			}
		}
	}
	if len(args) > 1 {
		code = args[1].(int)
	}
	ret := types.BaseRes{
		ApiCode:    code,
		ApiMessage: msg,
	}
	ctx.Resp.Json(ret)
}

// 业务成功返回资源
// args[0] message
// args[1] code
func SuccessResponse(ctx *httpserver.Context, args ...interface{}) {
	msg := "Successed"
	code := errcode.Success
	if len(args) > 0 {
		msg = args[0].(string)
	}
	if len(args) > 1 {
		code = args[1].(int)
	}
	ret := types.BaseRes{
		ApiCode:    code,
		ApiMessage: msg,
	}
	ctx.Resp.Json(ret)
}
