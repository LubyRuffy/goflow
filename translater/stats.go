package translater

import (
	"fmt"
	"github.com/LubyRuffy/goflow/workflowast"
)

// stats 指定字段统计
func statsHook(fi *workflowast.FuncInfo) string {
	// 调用zq
	field := ""
	if len(fi.Params) > 0 && len(fi.Params[0].RawString()) > 0 {
		field = "yield " + fi.Params[0].RawString() + " | "
	}

	size := ""
	if len(fi.Params) > 1 {
		size = fmt.Sprintf(" | tail %d", fi.Params[1].Int64())
	}
	return `ZqQuery(GetRunner(), map[string]interface{}{
    "query": "` + field + `sort | uniq -c | sort count` + size + `",
})`
}

func init() {
	register("stats", statsHook) // 统计
}
