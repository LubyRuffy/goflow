package goflow

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"sync"
)

// DumpTasks tasks dump to html
func (p *PipeRunner) DumpTasks(server bool, prefix string, fileMap sync.Map) string {
	t, err := template.New("tasks").Funcs(template.FuncMap{
		"toFileName": func(u string) string {
			return filepath.Base(u)
		},
		"HasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		"safeURL": func(u string, t string) template.URL {
			// 替换
			if v, ok := fileMap.Load(u); ok {
				return template.URL(v.(string))
			}

			if server {
				return template.URL(prefix + "/file?url=" + filepath.Base(u) + "&t=" + t)
			}
			u = strings.ReplaceAll(u, "\\", "/")
			return template.URL(u)
		},
		"GetTasks": func(p *PipeRunner) []*PipeTask {
			var ts []*PipeTask
			for _, wf := range p.GetWorkflows() {
				if wf.Runner != p {
					break
				}
				ts = append(ts, wf)
			}
			return ts
		},
	}).Parse(`

{{ template "task.tmpl" (GetTasks .) }}

{{ define "task.tmpl" }}
{{ range . }}
<ul>
	<li> {{ .Name }} ({{ .Content }}) </li>

	{{ if .Result }}

	{{ if gt (len .Result.OutFile) 0 }}
	<li><a href="{{ safeURL .Result.OutFile "" }}" target="_blank">{{ .Result.OutFile | toFileName }}</a></li>
	{{ end }}

	{{ if gt (len .Result.Artifacts) 0 }}
		<li>
		generate files:
		{{ range .Result.Artifacts }}
			<ul>
				<li><a href="{{ safeURL .FilePath .FileType  }}" target="_blank">
					{{ if HasPrefix .FileType "image/" }}
						<img src="{{ safeURL .FilePath .FileType }}" height="80px">
					{{ else if eq .FileType "chart_html"}}
						show <iframe width="660" height="520" src="{{ safeURL .FilePath .FileType }}" frameBorder="0"></iframe>
					{{ else }}
						{{ .FilePath | toFileName }}
					{{ end }}
				</a> | {{ .FileType }} | {{ .Memo }}</li>
			</ul>
		{{ end }}
		</li>
	{{ end }}
	{{ end }}

	<li>{{ .Cost }}</li>

	{{ range .Children }}
	<li> fork children:
		{{ template "task.tmpl" (GetTasks .) }}
	</li>
	{{ end }}
</ul>
{{ end }}
{{ end }}
`)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	err = t.Execute(&out, p)
	if err != nil {
		panic(err)
	}

	return out.String()
}
