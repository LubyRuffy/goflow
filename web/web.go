package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/LubyRuffy/goflow"
	"github.com/LubyRuffy/goflow/gocodefuncs"
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/gorilla/mux"
	"github.com/lubyruffy/gofofa"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"
)

//go:embed public
var webFs embed.FS
var (
	Prefix        string //
	getObjectHook = defaultGetObject
	newPipeRunner = defaultNewPipeRunner
)

func defaultGetObject(name string) (interface{}, bool) {
	return nil, false
}

// SetObjectHook 底层获取数据的回调接口
func SetObjectHook(f func(name string) (interface{}, bool)) {
	getObjectHook = f
}

func defaultNewPipeRunner() *goflow.PipeRunner {
	return goflow.New().WithDebug(logrus.DebugLevel)
}

// SetNewPipeRunner 创建Runner，上层可以自行注册自定义函数
func SetNewPipeRunner(f func() *goflow.PipeRunner) {
	newPipeRunner = f
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(webFs, "public/index.html"))
	tmpl.Execute(w, Prefix)
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
	return ast.ParseToGraph(code, func(name string, callId int, s string) string {
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

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

func create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	var a struct {
		Astcode string `json:"astcode"`
		Code    string `json:"code"`
	}

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    500,
			"message": fmt.Sprintf("workflow parsed err: %v", err),
		})
		return
	}

	var ast *workflowast.Parser
	var code string
	if a.Code != "" {
		code = a.Code
	} else if a.Astcode != "" {
		ast = workflowast.NewParser()
		code, err = ast.Parse(a.Astcode)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    500,
				"message": fmt.Sprintf("workflow parsed err: %v", err),
			})
			return
		}
	}

	tm := globalTaskMonitor.new(a.Astcode)

	hostinfo := "http://" + r.Host
	go func() {
		p := newPipeRunner().WithHooks(&goflow.Hooks{
			OnWorkflowStart: func(funcName string, actionId string) {
				tm.actionIDRunning = actionId
				tm.addMsg(fmt.Sprintf("workflow start: %s, %s", funcName, actionId))
			},
			OnWorkflowFinished: func(pt *goflow.PipeTask) {
				tm.addMsg(fmt.Sprintf("workflow finished: %s, %s", pt.Name, pt.ActionID))
			},
			OnLog: func(level logrus.Level, format string, args ...interface{}) {
				tm.addMsg(fmt.Sprintf("[%s] %s", level.String(), fmt.Sprintf(format, args...)))
			},
			OnGetObject: func(name string) (interface{}, bool) {
				v, ok := getObjectHook(name)
				if ok {
					return v, ok
				}
				switch name {
				case gocodefuncs.FofaObjectName:
					fofaCli, err := gofofa.NewClient()
					if err != nil {
						panic(fmt.Errorf("fofa connect err: %w", err))
					}
					return fofaCli, true
				}
				return nil, false
			},
		})

		tm.runner = p
		_, err = p.Run(code)
		if err != nil {
			tm.addMsg("create err: " + err.Error())
		}

		tm.html = p.DumpTasks(true, hostinfo+Prefix)
		tm.addMsg("<finished>")
		tm.finish()
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 200,
		"data": map[string]interface{}{
			"workflowId": tm.taskId,
		},
	})
}

func view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	var a struct {
		WorkflowId string `json:"workflowId"`
		TimeStamp  string `json:"timeStamp"`
	}

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    500,
			"message": fmt.Sprintf("workflow parsed err: %v", err),
		})
		return
	}

	t, ok := globalTaskMonitor.m.Load(a.WorkflowId)
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    500,
			"message": fmt.Sprintf("no task found"),
		})
		return
	}
	task := t.(*taskInfo)

	workflowStatus := "1"
	if task.finished {
		workflowStatus = "2"
	}

	msgs, ts := task.receiveMsgs(a.TimeStamp)

	returnObj := map[string]interface{}{
		"code":    200,
		"message": "success",
		"data": map[string]interface{}{
			"cost":           time.Since(task.started).String(),
			"workflowId":     a.WorkflowId,
			"timeStamp":      ts,
			"workflowStatus": workflowStatus,
			"logs":           msgs,
			"html":           task.html,
			"finished":       task.runner.Tasks,
			"active":         task.actionIDRunning,
		},
	}

	if len(task.astCode) > 0 {
		// ast 的运行模式
		ast := workflowast.NewParser()
		graphCode, err := genMermaidCode(ast, task.astCode)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    500,
				"message": fmt.Sprintf("workflow parsed err: %v", err),
			})
			return
		}

		if len(task.actionIDRunning) > 0 {
			graphCode += fmt.Sprintf("\nstyle F%s fill:#57d3e3", task.actionIDRunning)
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
				graphCode += fmt.Sprintf("\nstyle F%s fill:%s", ti.ActionID, color)

			}
		}

		returnObj["data"].(map[string]interface{})["graphCode"] = graphCode
	}

	json.NewEncoder(w).Encode(returnObj)
}

// AppendToMuxRouter 附加到路由注册中
func AppendToMuxRouter(router *mux.Router) {
	// 静态资源
	router.PathPrefix(Prefix + "/public/").Handler(
		http.StripPrefix(Prefix+"/",
			http.FileServer(http.FS(webFs)),
		),
	)

	// 默认首页
	router.HandleFunc(Prefix+"/", handler)

	// 任务
	router.HandleFunc(Prefix+"/parse", parse)
	router.HandleFunc(Prefix+"/api/v1/workflow/create", create)
	router.HandleFunc(Prefix+"/api/v1/workflow/view", view)
	router.HandleFunc(Prefix+"/file", func(w http.ResponseWriter, r *http.Request) {
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
}

// Start 启动服务器
func Start(addr string) error {
	router := mux.NewRouter()
	AppendToMuxRouter(router)

	logrus.Println("listen at: ", addr)
	return http.ListenAndServe(addr, router)
}
