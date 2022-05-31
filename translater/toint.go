package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
)

func intHook(fi *workflowast.FuncInfo) string {
	// 调用zq
	return `ZqQuery(GetRunner(), map[string]interface{}{
    "query": "cast(this, <{` + fi.Params[0].RawString() + `:int64}>) ",
})`
}

func init() {
	Register("to_int", intHook) // 将某个字段转换为int类型
}
