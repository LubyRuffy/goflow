package utils

import (
	"runtime"
	"strings"
)

// FunctionName 当前调用的函数名称
func FunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc).Name()
	// 完整的路径是这样的：github.com/LubyRuffy/goflow/utils.TestFunctionName
	fn = fn[strings.LastIndex(fn, "/")+1:] //去掉前面所有的部分
	if fs := strings.Split(fn, "."); len(fs) > 1 {
		fn = fs[1]
	}
	return fn
}
