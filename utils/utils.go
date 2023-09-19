package utils

import (
	"fmt"
	"runtime"
	"strings"
)

// 显示错误时运行时的堆栈信息
func Trace(errMessage string) string {
	var pcstack [32]uintptr
	// Callers 将调用函数的返回程序计数器填入调用 goroutine 堆栈的片段 pc。
	// 参数 skip 是在 pc 中记录之前要跳过的堆栈帧数，0 表示 Callers 本身的帧，1 表示 Callers 的调用者。它返回写入 pc 的条目数。
	n := runtime.Callers(3, pcstack[:])

	// Using Builder optimize speed.
	var str strings.Builder
	str.WriteString(errMessage + "\nTraceback:")
	for _, pc := range pcstack[:n] {
		function := runtime.FuncForPC(pc)
		file, line := function.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

// 判断是否满足 ip:port 的格式
func ValidPerrAddr(addr string) bool {
	token1 := strings.Split(addr, ":")
	if len(token1) != 2 {
		return false
	}
	token2 := strings.Split(token1[0], ".")
	if token2[0] != "localhost" && len(token2[0]) != 4 {
		return false
	}
	return true
}
