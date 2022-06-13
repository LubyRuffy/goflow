// Package translater workflow to gocode
package translater

import (
	"fmt"
	"github.com/LubyRuffy/goflow/workflowast"
	"runtime"
	"strings"
)

var (
	Translators []string
)

// Register 注册翻译函数，从workflow的函数展开为底层gocode的完整代码
func Register(name string, f workflowast.FunctionTranslateHook) {
	Translators = append(Translators, name)
	workflowast.RegisterFunction(name, f)
}

// diePanic 打印带有函数名称的奔溃，参考utils.FunctionName()
func diePanic(err error) {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc).Name()
	fn = fn[strings.LastIndex(fn, "/")+1:] //去掉前面所有的部分
	panic(fmt.Errorf(fn+" failed: %w", err))
}
