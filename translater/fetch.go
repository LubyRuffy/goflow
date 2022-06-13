package translater

import (
	"bytes"
	"github.com/LubyRuffy/goflow/workflowast"
	"text/template"
)

func fetchHook(fi *workflowast.FuncInfo) string {
	tmpl, err := template.New("load").Parse(`FetchFile(GetRunner(), map[string]interface{} {
    "url": {{ .URL }},
})`)
	if err != nil {
		diePanic(err)
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, struct {
		URL string
	}{
		URL: fi.Params[0].String(),
	})
	if err != nil {
		diePanic(err)
	}
	return tpl.String()
}

func init() {
	Register("fetch", fetchHook)
}
