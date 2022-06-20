package translater

import "github.com/LubyRuffy/goflow/workflowast"

func whereHook(fi *workflowast.FuncInfo) string {
	// 调用zq
	return `ZqQuery(GetRunner(), map[string]interface{}{
    "query": "where ` + fi.Params[0].RawString() + `",
})`
}

func init() {
	Register("where", whereHook) // 将某个字段转换为int类型
}
