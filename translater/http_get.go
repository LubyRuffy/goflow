package translater

import (
	"bytes"
	"github.com/LubyRuffy/goflow/workflowast"
	"text/template"
)

func httpGetHook(fi *workflowast.FuncInfo) string {
	tmpl, err := template.New("http_get").Parse(`HttpRequest(GetRunner(), map[string]interface{} {
    "urlField": "{{ .URLField }}",
    "userAgent": "{{ .UserAgent }}",
    "tlsVerify": {{ .TLSVerify }},
    "workers": {{ .Workers }},
    "maxSize": {{ .MaxSize }},
})`)
	if err != nil {
		panic(err)
	}

	urlField := "url"
	if len(fi.Params) > 0 {
		urlField = fi.Params[0].RawString()
	}
	userAgent := ""
	if len(fi.Params) > 1 {
		userAgent = fi.Params[1].RawString()
	}
	tlsVerify := "false"
	if len(fi.Params) > 2 {
		tlsVerify = fi.Params[2].ToString()
	}
	workers := 5
	if len(fi.Params) > 3 {
		workers = int(fi.Params[3].Int64())
	}
	maxSize := -1
	if len(fi.Params) > 4 {
		maxSize = int(fi.Params[4].Int64())
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, struct {
		URLField  string
		UserAgent string
		TLSVerify string
		Workers   int
		MaxSize   int
	}{
		URLField:  urlField,
		UserAgent: userAgent,
		TLSVerify: tlsVerify,
		Workers:   workers,
		MaxSize:   maxSize,
	})
	if err != nil {
		panic(err)
	}
	return tpl.String()
}

func init() {
	Register("http_get", httpGetHook)
}
