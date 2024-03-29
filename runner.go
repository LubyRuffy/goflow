package goflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/traefik/yaegi/interp"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/LubyRuffy/goflow/gocodefuncs"

	"github.com/LubyRuffy/goflow/coderunner"
	"github.com/LubyRuffy/goflow/translater"
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/sirupsen/logrus"
)

// PipeTask 每一个pipe执行的任务统计信息
type PipeTask struct {
	Name     string                  `json:"name"`     // pipe name
	Content  string                  `json:"-"`        // raw content
	Runner   *PipeRunner             `json:"-"`        // runner
	ActionID string                  `json:"actionId"` // 调用序列
	Cost     time.Duration           `json:"cost"`     // time costs
	Result   *gocodefuncs.FuncResult `json:"result"`   // 结果
	Children []*PipeRunner           `json:"-"`        // fork children
	Fields   []string                `json:"fields"`   // fields list 列名
	Error    string                  `json:"error"`    // 错误信息
}

// Close remove tmp outfile
func (p *PipeTask) Close() {
	os.Remove(p.Result.OutFile)
	for _, f := range p.Result.Artifacts {
		os.Remove(f.FilePath)
	}
	p.Children = nil
}

// PipeRunner pipe运行器
type PipeRunner struct {
	ctx          context.Context        // 运行上下文，run时才进行设置
	cancel       context.CancelFunc     // 中止运行的函数
	gf           *coderunner.GoFunction // 函数注册
	ast          *workflowast.Parser    // ast
	content      string                 // 运行的内容
	hooks        *Hooks                 // 消息通知
	Tasks        []*PipeTask            // 执行的所有workflow
	LastTask     *PipeTask              // 最后执行的workflow
	LastFile     string                 // 最后生成的文件名
	logger       *logrus.Logger
	children     []*PipeRunner
	Parent       *PipeRunner
	gocodeRunner *coderunner.Runner // 底层的代码执行器
	objects      sync.Map           // 全局注册的对象
	WebHook      string             // webhook对应的地址
	HasResult    bool               // 是否存在最终结果，用于sdk模式下的扣费辅助判断
}

// GetObject 获取全局变量
func (p *PipeRunner) GetObject(name string) (interface{}, bool) {
	if v, ok := p.hooks.OnGetObject(name); ok {
		return v, ok
	}
	return p.objects.Load(name)
}

// SetObject 获取全局变量
func (p *PipeRunner) SetObject(name string, value interface{}) {
	p.objects.Store(name, value)
}

// Logf 打印日志
func (p *PipeRunner) Logf(level logrus.Level, format string, args ...interface{}) {
	p.logger.Logf(level, format, args...)
}

// Debugf 打印调试日志
func (p *PipeRunner) Debugf(format string, args ...interface{}) {
	p.Logf(logrus.DebugLevel, format, args...)
}

// Warnf 打印警告日志
func (p *PipeRunner) Warnf(format string, args ...interface{}) {
	p.Logf(logrus.WarnLevel, format, args...)
}

// doWebHook
func (p *PipeRunner) doWebHook(info map[string]interface{}) {
	if p.WebHook != "" {
		d, err := json.Marshal(info)
		if err != nil {
			p.logger.Errorf("doWebHook data failed: %v", err)
		}
		resp, err := http.Post(p.WebHook, "text/json", bytes.NewReader(d))
		if err != nil {
			p.logger.Errorf("doWebHook post failed: %v", err)
		}
		p.logger.Debugf("doWebHook response code: %d", resp.StatusCode)
	}
}

// Run go code, not workflow
func (p *PipeRunner) Run(ctx context.Context, code string) (reflect.Value, error) {
	s := time.Now()
	p.content = code
	p.ctx, p.cancel = context.WithCancel(ctx)
	v, err := p.gocodeRunner.Run(code)
	if err != nil {
		// 这里打印错误信息
		if err1, ok := err.(interp.Panic); ok {
			log.Println("run code failed with:\n", string(err1.Stack))
		} else {
			log.Println(err)
		}
	}
	p.HasResult = !p.LastFileEmpty()
	p.doWebHook(map[string]interface{}{
		"event": "finished",
		"cost":  time.Since(s).Truncate(time.Millisecond).String(),
		"tasks": p.Tasks,
	})
	return v, err
}

// Close remove tmp outfile
func (p *PipeRunner) Close() {
	p.children = nil
	p.LastFile = ""
	p.LastTask = nil
	for _, task := range p.Tasks {
		task.Close()
	}
	p.Tasks = nil
}

// GetWorkflows all workflows
func (p *PipeRunner) GetWorkflows() []*PipeTask {
	return p.Tasks
}

// AddWorkflow 添加一次任务的日志
func (p *PipeRunner) AddWorkflow(pt *PipeTask) {
	// 可以不写文件
	if pt.Result != nil && len(pt.Result.OutFile) > 0 {
		p.LastFile = pt.Result.OutFile

		logrus.Info(pt.Name+" write to file: ", pt.Result.OutFile)

		// 取字段列表
		pt.Fields = utils.JSONLineFields(pt.Result.OutFile)
	}
	p.LastTask = pt

	// 把任务也加到上层所有的父节点
	node := p
	for {
		node.Tasks = append(node.Tasks, pt)
		if node.Parent == nil {
			break
		}
		node = node.Parent
	}

	if p.hooks != nil {
		if pt.Error != "" {
			p.hooks.OnLog(logrus.ErrorLevel, "task error: %s", pt.Error)
		}
		p.hooks.OnWorkflowFinished(pt)
	}
}

// GetLastFile last genrated file
func (p *PipeRunner) GetLastFile() string {
	return p.LastFile
}

// 核心函数
func (p *PipeRunner) fork(pipe string) error {
	forkRunner := New().WithHooks(p.hooks).WithParent(p)
	forkRunner.LastFile = p.LastFile // 从这里开始分叉
	p.LastTask.Children = append(p.LastTask.Children, forkRunner)
	code, err := workflowast.NewParser().Parse(pipe)
	if err != nil {
		return err
	}
	p.children = append(p.children, forkRunner)
	newCtx, cancel := context.WithCancel(p.ctx)
	defer cancel()
	_, err = forkRunner.Run(newCtx, code)
	return err
}

// registerFunctions 注册用户自定义函数，做一层PipeTask封装
// func为可选长度，如果是两个，说明只注册底层函数，如果是4个，说明要注册翻译函数，让workflow使用
func (p *PipeRunner) registerFunctions(funcs ...[]interface{}) {
	for i := range funcs {
		funcName := funcs[i][0].(string)
		funcBody := funcs[i][1].(func(gocodefuncs.Runner, map[string]interface{}) *gocodefuncs.FuncResult)

		if len(funcs[i]) > 3 {
			translater.Register(funcs[i][2].(string),
				funcs[i][3].(func(fi *workflowast.FuncInfo) string))
		}

		p.gf.Register(funcName, func(runner gocodefuncs.Runner, params map[string]interface{}) {
			var actionId string
			if v, ok := params["actionId"]; ok {
				actionId = v.(string)
			} else {
				callID := 1
				node := p
				for {
					callID = len(node.Tasks) + 1
					if node.Parent == nil {
						break
					}
					node = node.Parent
				}
				actionId = strconv.Itoa(callID)
			}

			// todo: 可以在这里格式化参数，展开为全局变量，但是这样会导致全局变量优先级高于运行时变量，同时并不能解决每个块执行时需要初始化运行时变量的问题

			logrus.Debug(funcName+" params:", params)
			if p.hooks != nil {
				p.hooks.OnWorkflowStart(funcName, actionId)
			}

			s := time.Now()
			pt := &PipeTask{
				Name:     funcName,
				Content:  fmt.Sprintf("%v", params),
				ActionID: actionId,
				Runner:   p,
			}

			// 异常捕获
			defer func() {
				if r := recover(); r != nil {
					pt.Error = r.(error).Error()
					pt.Cost = time.Since(s)
					p.AddWorkflow(pt)
					p.Logf(logrus.ErrorLevel,
						"action %s running with param %s failed, error: %s",
						pt.ActionID, pt.Content, r.(error).Error())
					panic(r.(error).Error())
				}
			}()

			result := funcBody(p, params)
			pt.Result = result
			pt.Cost = time.Since(s)

			p.AddWorkflow(pt)
		})
	}
}

// TarFinalOutputs 打包最终输出的json与对应的资源文件
func (p *PipeRunner) TarFinalOutputs(tarFunc func(files []string) ([]byte, error)) ([]byte, error) {
	var files []string
	var err error

	// 获取需要加载文件的字段
	if flds, ok := p.GetObject(utils.ResourceFieldsObjectName); ok {
		// 获取文件资源
		if resources, ok := p.GetObject(utils.ResourcesObjectName); ok {
			err = utils.EachLine(p.LastFile, func(line string) error {
				for _, fld := range flds.([]string) {
					file := gjson.Get(line, fld).String()
					// 判断当前文件是否在生成文件列表中
					if file != "" && utils.Contains(resources.([]string), file) {
						files = append(files, file)
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		}

	} else {
		// 没有需要打包的字段
	}

	// 获取static文件(不存在于json文件里，也不会被修改)
	if resources, ok := p.GetObject(utils.StaticResourceObjectName); ok {
		for _, resource := range resources.([]string) {
			files = append(files, resource)
		}
	}

	// 格式化最终输出的json
	fn, err := p.FormatResourceFieldInJson(p.LastFile)
	if err != nil {
		return nil, err
	}
	files = append(files, fn)

	// 文件打包
	if len(files) > 0 {
		tarGzData, err := tarFunc(files)
		if err != nil {
			return nil, err
		}
		return tarGzData, nil
	}

	return nil, nil
}

// FormatResourceFieldInJson 将资源文件对应字段中的内容从绝对路径改为文件名
func (p *PipeRunner) FormatResourceFieldInJson(filename string) (fn string, err error) {
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		err = utils.EachLine(filename, func(line string) error {
			// 逐行进行文件名替换
			if flds, ok := p.GetObject(utils.ResourceFieldsObjectName); ok {
				for _, fld := range flds.([]string) {
					file := gjson.Get(line, fld).String()
					if file != "" {
						// 此处存在字段被删除的情况，需要在字段没有被删除的情况下再进行文件替换
						line, _ = sjson.Set(line, fld, filepath.Base(file))
					}
				}
			}
			f.WriteString(line + "\n")
			return nil
		})
		return err
	})
	return
}

func (p *PipeRunner) OnJobStart() {
	return
}

func (p *PipeRunner) OnJobFinished() {
	return
}

// LastFileEmpty 判断最终文件是否为空
func (p *PipeRunner) LastFileEmpty() bool {
	info, err := os.Stat(p.LastFile)
	if err != nil {
		panic(err)
	}
	return info.Size() == 0
}

// TarGzAll 打包所有文件
func (p *PipeRunner) TarGzAll(tarFunc func(files []string) ([]byte, error)) ([]byte, error) {
	var files []string
	for _, t := range p.Tasks {
		if t.Result == nil {
			continue
		}

		if len(t.Result.OutFile) > 0 {
			files = append(files, t.Result.OutFile)
		}
		if len(t.Result.Artifacts) > 0 {
			for _, f := range t.Result.Artifacts {
				files = append(files, f.FilePath)
			}
		}
	}

	if len(files) > 0 {
		tarGzData, err := tarFunc(files)
		if err != nil {
			return nil, err
		}
		return tarGzData, nil
	}

	return nil, nil
}

// SetProgress 设置进度
func (p *PipeRunner) SetProgress(v float64) {
	if p.hooks != nil && p.hooks.OnProgress != nil {
		p.hooks.OnProgress(100 * v)
	}
	p.logger.Printf("progress: %f%%", 100*v)
}

// GetContext 获取ctx
func (p *PipeRunner) GetContext() context.Context {
	return p.ctx
}

// Stop 停止运行
func (p *PipeRunner) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

// logHook is a hook designed for dealing with logs in test scenarios.
type logHook struct {
	runner *PipeRunner
	mu     sync.RWMutex
}

func (t *logHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t *logHook) Fire(e *logrus.Entry) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.runner != nil && t.runner.hooks != nil && t.runner.hooks.OnLog != nil {
		t.runner.hooks.OnLog(e.Level, e.Message)
	}
	return nil
}

// New create pipe runner
func New() *PipeRunner {
	r := &PipeRunner{
		logger: logrus.New(),
		gf:     &coderunner.GoFunction{},
		hooks:  defaultHooks,
	}
	r.logger.AddHook(&logHook{runner: r}) // fofa日志打到前端

	var err error

	// 注册底层函数
	err = r.gf.Register("GetRunner", func() *PipeRunner {
		return r
	})
	if err != nil {
		panic(err)
	}
	err = r.gf.Register("Fork", r.fork)
	if err != nil {
		panic(err)
	}
	err = r.gf.Register("ZqValue", gocodefuncs.ZqValue) // 取zq计算后的值
	if err != nil {
		panic(err)
	}

	innerFuncs := [][]interface{}{
		{"RemoveField", gocodefuncs.RemoveField},
		{"FetchFofa", gocodefuncs.FetchFofa},
		{"FetchFile", gocodefuncs.FetchFile},
		{"GenFofaFieldData", gocodefuncs.GenFofaFieldData},
		{"GenerateChart", gocodefuncs.GenerateChart},
		{"PieChart", gocodefuncs.PieChart},
		{"BarChart", gocodefuncs.BarChart},
		{"ZqQuery", gocodefuncs.ZqQuery},
		{"AddField", gocodefuncs.AddField},
		{"LoadFile", gocodefuncs.LoadFile},
		{"FlatArray", gocodefuncs.FlatArray},
		//{"Screenshot", gocodefuncs.Screenshot},	// move to agent
		{"ToExcel", gocodefuncs.ToExcel},
		{"ToSql", gocodefuncs.ToSql},
		{"GenData", gocodefuncs.GenData},
		{"URLFix", gocodefuncs.UrlFix},
		//{"RenderDOM", gocodefuncs.RenderDOM},
		{"ScanPort", gocodefuncs.ScanPort},
		{"ParseURL", gocodefuncs.ParseURL},
		{"HttpRequest", gocodefuncs.HttpRequest},
		{"TextClassify", gocodefuncs.TextClassify},
		{"JoinFofa", gocodefuncs.JoinFofa},
		{"Merge", gocodefuncs.Merge},
		{"ExcelToJson", gocodefuncs.ExcelToJson},
		{"CSVToJson", gocodefuncs.CSVToJson},
		{"ZipToJson", gocodefuncs.ZipToJson},
		{"JqQuery", gocodefuncs.JqQuery},
		{"Join", gocodefuncs.Join},
	}
	r.registerFunctions(innerFuncs...)

	logrus.Debug("ast support workflows:", translater.Translators)

	r.gocodeRunner = coderunner.New(coderunner.WithFunctions(r.gf))

	return r
}
