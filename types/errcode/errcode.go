package errcode

const Success = 0

const Fail = 1

const InvalidToken = 1000

const NoPermission = 1100

const LoginExpire = 1200
const LoginAccountPasswordFail = 1201 // 账号或密码不对

const TelExists = 2000            // 注册检查电话已存在
const VerifyCodeError = 2001      // 验证码错误
const InviteCodeError = 2002      // 邀请码错误
const VerifyCodeFrequently = 2003 // 获取验证码频繁
const GlobeSmsSendtFail = 2004    // 初发送国际短信失败，可能欠费了或者不支持该国家
