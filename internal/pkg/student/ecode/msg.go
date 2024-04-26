package ecode

var MsgFlags = map[int]string{
	SUCCESS: "ok",
	ERROR:   "fail",

	InvalidParameters:          "请求参数错误",
	HaveSignUp:                 "已经报名了",
	ErrorActivityTimeout:       "活动过期了",
	ErrorAuthCheckTokenFail:    "Token鉴权失败",
	ErrorAuthCheckTokenTimeout: "Token已超时",
	ErrorAuthToken:             "Token生成失败",
	ErrorAuth:                  "Token错误",
	ErrorNotCompare:            "不匹配",
	ErrorDatabase:              "数据库操作出错,请重试",
	ErrorAuthNotFound:          "Token不能为空",

	ErrorServiceUnavailable: "过载保护，服务暂时不可用",
	ErrorDeadline:           "服务调用超时",
}

func GetMsg(code int) string {
	if msg, ok := MsgFlags[code]; ok {
		return msg
	}
	return MsgFlags[ERROR]
}
