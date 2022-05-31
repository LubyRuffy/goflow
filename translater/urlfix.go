package translater

import "github.com/LubyRuffy/goflow/workflowast"

func urlfixHook(fi *workflowast.FuncInfo) string {
	urlField := "url"
	if len(fi.Params) > 0 {
		urlField = fi.Params[0].RawString()
	}
	return `URLFix(GetRunner(), map[string]interface{}{
    "urlField": "` + urlField + `",
})`
}

func init() {
	Register("urlfix", urlfixHook) // 补充完善url
}
