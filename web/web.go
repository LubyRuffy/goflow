package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/lubyruffy/gofofa"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/LubyRuffy/goflow"
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/sirupsen/logrus"
)

//go:embed public
var webFs embed.FS

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(webFs, "public/index.html"))
	tmpl.Execute(w, "")
}

func genMermaidCode(ast *workflowast.Parser, code string) (s string, err error) {
	// 输入源
	sourceWorkflow := []string{
		"load", "fofa", "scan_port",
	}
	// 终止
	finishWorkflow := []string{
		"chart", "to_excel", "to_mysql", "to_sqlite",
	}
	return ast.ParseToGraph(string(code), func(name string, callId int, s string) string {
		for _, src := range sourceWorkflow {
			if src == name {
				return `F` + strconv.Itoa(callId) + `[("` + s + `")]`
			}
		}
		for _, src := range finishWorkflow {
			if src == name {
				return `F` + strconv.Itoa(callId) + `[["` + s + `"]]`
			}
		}
		return `F` + strconv.Itoa(callId) + `["` + s + `"]`
	}, "graph LR\n")
}

func parse(w http.ResponseWriter, r *http.Request) {
	// fofa(`title=test`) & to_int(`port`) & sort(`port`) & [cut(`port`) | cut("ip")]
	w.Header().Set("Content-Type", "application/json")

	code, err := ioutil.ReadAll(r.Body)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  true,
			"result": fmt.Sprintf("workflow parsed err: %v", err),
		})
		return
	}

	ast := workflowast.NewParser()
	realCode, err := ast.Parse(string(code))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  true,
			"result": fmt.Sprintf("workflow parsed err: %v", err),
		})
		return
	}
	var calls []string
	for _, fi := range ast.CallList {
		calls = append(calls, fi.Name)
	}

	graphCode, err := genMermaidCode(ast, string(code))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  true,
			"result": fmt.Sprintf("workflow parsed err: %v", err),
		})
		return
	}
	logrus.Println(graphCode)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": false,
		"result": map[string]interface{}{
			"realCode":  realCode,
			"graphCode": graphCode,
			"calls":     calls,
		},
	})
}

func run(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	workflow, err := ioutil.ReadAll(r.Body)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  true,
			"result": fmt.Sprintf("workflow parsed err: %v", err),
		})
		return
	}

	var code string
	ast := workflowast.NewParser()
	code, err = ast.Parse(string(workflow))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  true,
			"result": fmt.Sprintf("workflow parsed err: %v", err),
		})
		return
	}

	tm := globalTaskMonitor.new(string(workflow))

	go func() {
		fofaCli, err := gofofa.NewClient()
		if err != nil {
			panic(fmt.Errorf("fofa connect err: %w", err))
		}

		p := goflow.New(goflow.WithAST(ast), goflow.WithObject("fofacli", fofaCli),
			goflow.WithHooks(&goflow.Hooks{
				OnWorkflowStart: func(funcName string, callID int) {
					tm.callIDRunning = callID
					tm.addMsg(fmt.Sprintf("workflow start: %s, %s, %d", ast.CallList[callID-1].Name, funcName, callID))
				},
				OnWorkflowFinished: func(pt *goflow.PipeTask) {
					tm.addMsg(fmt.Sprintf("workflow finished: %s, %s, %d", pt.WorkFlowName, pt.Name, pt.CallID))
				},
				OnLog: func(level logrus.Level, format string, args ...interface{}) {
					tm.addMsg(fmt.Sprintf("[%s] %s", level.String(), fmt.Sprintf(format, args...)))
				},
			}))
		tm.runner = p
		_, err = p.Run(code)
		if err != nil {
			tm.addMsg("run err: " + err.Error())
		}

		tm.html = p.DumpTasks(true)
		tm.addMsg("<finished>")
		tm.finish()
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  false,
		"result": tm.taskId,
	})
}

func fetchMsg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tid := r.FormValue("tid")

	t, ok := globalTaskMonitor.m.Load(tid)
	task := t.(*taskInfo)
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  true,
			"result": fmt.Sprintf("no task found"),
		})
		return
	}
	var msgs []string
	s := time.Now()
	for {
		log.Println(time.Since(s))
		info, ok := task.receiveMsg()
		if !ok {
			break
		}
		msgs = append(msgs, info)
	}

	ast := workflowast.NewParser()
	graphCode, err := genMermaidCode(ast, task.code)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  true,
			"result": fmt.Sprintf("workflow parsed err: %v", err),
		})
		return
	}

	if task.callIDRunning > 0 {
		graphCode += fmt.Sprintf("\nstyle F%d fill:#57d3e3", task.callIDRunning)
	}

	if task.runner != nil {
		for i := range task.runner.Tasks {
			ti := task.runner.Tasks[i]
			color := ""
			if ti.Error != nil {
				color = "#fc8fa3"
			} else {
				color = "#65d9ae"
			}
			graphCode += fmt.Sprintf("\nstyle F%d fill:%s", ti.CallID, color)

		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": false,
		"result": map[string]interface{}{
			"msgs":      msgs,
			"html":      task.html,
			"graphCode": graphCode,
		},
	})
}

// Start 启动服务器
func Start(addr string) error {
	// 默认首页
	http.HandleFunc("/", handler)
	http.HandleFunc("/parse", parse)
	http.HandleFunc("/run", run)
	http.HandleFunc("/fetchMsg", fetchMsg)
	http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		fn := filepath.Base(r.FormValue("url"))
		f := filepath.Join(os.TempDir(), fn)
		needRawFilename := false
		switch filepath.Ext(fn) {
		case ".sql", ".xlsx", ".sqlite3":
			needRawFilename = true
		}
		if len(r.FormValue("dl")) > 0 {
			needRawFilename = true
		}
		if needRawFilename {
			w.Header().Set("Content-Disposition", "attachment; filename="+fn)
		}
		http.ServeFile(w, r, f)
	})

	// 静态资源
	http.Handle("/public/", http.StripPrefix("/",
		http.FileServer(http.FS(webFs))))

	logrus.Println("listen at: ", addr)
	return http.ListenAndServe(addr, nil)
}