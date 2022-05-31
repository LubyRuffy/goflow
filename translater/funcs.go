// Package translater workflow to gocode
package translater

import "github.com/LubyRuffy/goflow/workflowast"

var (
	Translators []string
)

// Register 注册翻译函数，从workflow的函数展开为底层gocode的完整代码
func Register(name string, f workflowast.FunctionTranslateHook) {
	Translators = append(Translators, name)
	workflowast.RegisterFunction(name, f)
}
