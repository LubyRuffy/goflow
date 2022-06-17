package translater

import (
	"bytes"
	"github.com/LubyRuffy/goflow/workflowast"
	"text/template"
)

func fakefofaHook(fi *workflowast.FuncInfo) string {
	tmpl, err := template.New("fakefofa").Parse(`GenFofaFieldData(GetRunner(), map[string]interface{} {
    "query": {{ .Query }},
    "size": {{ .Size }},
    "fields": {{ .Fields }},
})`)
	if err != nil {
		panic(err)
	}
	var size int64 = 10
	fields := "`host,title,ip,port`"
	if len(fi.Params) > 1 {
		fields = fi.Params[1].String()
	}
	if len(fi.Params) > 2 {
		size = fi.Params[2].Int64()
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, struct {
		Query  string
		Size   int64
		Fields string
	}{
		Query:  fi.Params[0].String(),
		Fields: fields,
		Size:   size,
	})
	if err != nil {
		panic(err)
	}
	return tpl.String()
}

func init() {
	Register("fakefofa", fakefofaHook)
}
