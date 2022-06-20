package translater

import "github.com/LubyRuffy/goflow/workflowast"

func parseUrlHook(fi *workflowast.FuncInfo) string {
	// 调用zq
	urlField := "url"
	if len(fi.Params) > 0 {
		urlField = fi.Params[0].RawString()
	}
	return `ZqQuery(GetRunner(), map[string]interface{}{
	"query": "yield {...this, ` + urlField + `_parsed:parse_uri(` + urlField + `)}",
})`
}

func init() {
	Register("parse_url", parseUrlHook) // 将某个字段转换为int类型
}
