package translater

import (
	"bytes"
	"github.com/LubyRuffy/goflow/workflowast"
	"text/template"
)

// bar 生成bar报表
// bar("name_field","value_field", size, "chart title")
// 第一个参数是显示的字段名称；第二个参数是值的字段名称，如果是count()表明是去重统计；第三个参数是top size；第四个参数是标题
func bar(fi *workflowast.FuncInfo) string {
	tmpl, _ := template.New("pie").Parse(`BarChart(GetRunner(), map[string]interface{}{
    "name": "{{ .Name }}",
    "value": "{{ .Value }}",
    "size": {{ .Size }},
    "title": "{{ .Title }}",
})`)

	value := "count()"
	if len(fi.Params) > 1 {
		if v := fi.Params[1].RawString(); len(v) > 0 {
			value = fi.Params[1].RawString()
		}
	}
	size := -1
	if len(fi.Params) > 2 {
		size = int(fi.Params[2].Int64())
	}
	title := ""
	if len(fi.Params) > 3 {
		title = fi.Params[3].RawString()
	}

	var tpl bytes.Buffer
	err := tmpl.Execute(&tpl, struct {
		Name  string
		Value string
		Size  int
		Title string
	}{
		Name:  fi.Params[0].RawString(),
		Value: value,
		Size:  size,
		Title: title,
	})
	if err != nil {
		panic(err)
	}
	return tpl.String()
}

func init() {
	Register("bar", bar)
}
