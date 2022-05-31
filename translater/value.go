package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
)

func valueHook(fi *workflowast.FuncInfo) string {
	// 调用zq
	return `ZqQuery(GetRunner(), map[string]interface{}{
    "query": "yield ` + fi.Params[0].RawString() + `",
})`
}

func init() {
	Register("value", valueHook) // 取值
}
