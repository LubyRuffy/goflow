package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
)

func dropHook(fi *workflowast.FuncInfo) string {
	// 调用zq
	return `ZqQuery(GetRunner(), map[string]interface{}{
    "query": "drop ` + fi.Params[0].RawString() + `",
})`
}

func init() {
	register("drop", dropHook) // 删除字段
}
