package translater

import "github.com/LubyRuffy/goflow/workflowast"

func fixURLHook(fi *workflowast.FuncInfo) string {
	urlField := "url"
	if len(fi.Params) > 0 {
		urlField = fi.Params[0].RawString()
	}
	return `URLFix(GetRunner(), map[string]interface{}{
    "urlField": "` + urlField + `",
})`
}

func init() {
	Register("fix_url", fixURLHook) // 补充完善url
}
