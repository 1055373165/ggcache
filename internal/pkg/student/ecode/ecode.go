package ecode

const (
	SUCCESS           = 200
	ERROR             = 500
	InvalidParameters = 400

	ErrorExistUser      = 10001
	ErrorNotExistUser   = 10002
	ErrorFailEncryption = 10003
	ErrorNotCompare     = 10004

	HaveSignUp           = 20001
	ErrorActivityTimeout = 20002

	ErrorAuthCheckTokenFail    = 30001
	ErrorAuthCheckTokenTimeout = 30002
	ErrorAuthToken             = 30003
	ErrorAuth                  = 30004
	ErrorAuthNotFound          = 30005

	ErrorDatabase = 40001

	ErrorServiceUnavailable = 50001
	ErrorDeadline           = 50002
)
