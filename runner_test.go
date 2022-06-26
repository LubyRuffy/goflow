package goflow

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"database/sql"
	"github.com/LubyRuffy/goflow/gocodefuncs"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/xuri/excelize/v2"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/LubyRuffy/gofofa"
	"github.com/stretchr/testify/assert"
)

// 返回值表明是错误并且匹配到了错误返回
func defaultErrorHandler(t *testing.T, err error) bool {
	assert.Nil(t, err)
	return false
}

func assertPipeCmd(t *testing.T, workflow string, testData string,
	options ...interface{}) string {
	var err error
	r := New()

	errHandler := defaultErrorHandler
	if len(options) > 0 {
		errHandler = options[0].(func(t *testing.T, err error) bool)
	}

	// 写入数据文件
	if len(testData) > 0 {
		r.LastFile, err = utils.WriteTempFile("", func(f *os.File) error {
			_, err = f.WriteString(testData)
			return err
		})
		if errHandler(t, err) {
			return ""
		}
	}

	// 执行代码
	code, err := workflowast.NewParser().Parse(workflow)
	if errHandler(t, err) {
		return ""
	}

	_, err = r.Run(code)
	if errHandler(t, err) {
		return ""
	}

	var data []byte
	data, err = os.ReadFile(r.GetLastFile())
	if errHandler(t, err) {
		return ""
	}
	return string(data)
}

func assertPipeCmdByTestRunner(t *testing.T, workflow string, testData string,
	except string, options ...interface{}) {
	data := assertPipeCmd(t, workflow, testData, options...)
	assert.Equal(t, except, data)
}

func assertPipeCmdByTestRunnerError(t *testing.T,
	workflow string, testData string, errorStr string) {
	assertPipeCmdByTestRunner(t, workflow, testData, "", func(t *testing.T, err error) bool {
		if err != nil {
			assert.Contains(t, err.Error(), errorStr)
			return true
		}
		return false
	})
}

func TestNew_cut(t *testing.T) {
	assertPipeCmdByTestRunner(t, `cut("a")`, `{"a":1,"b":2}`, "{\"a\":1}\n")
	//assertPipeCmd(t, `cut("a")`, `{"a":1,"b":2}`, "{\"a\":1}\n")
}

func TestNew(t *testing.T) {
	assertPipeCmdByTestRunner(t, `add("newfield", "newvalue")`,
		`{"title":"Test123"}
{"title":"123test456"}`,
		`{"title":"Test123","newfield":"newvalue"}
{"title":"123test456","newfield":"newvalue"}`)

	// chart格式错误
	assertPipeCmdByTestRunnerError(t, `chart("line","a")`,
		`{"title":"Test123"}`,
		`"value" and "count" field is needed`)
	// chart正确
	assertPipeCmdByTestRunner(t, `chart("bar","a")`,
		`{"value":"Test123","count":10}`,
		"{\"value\":\"Test123\",\"count\":10}")
	// chart正确
	assertPipeCmdByTestRunner(t, `chart("pie","a")`,
		`{"value":"Test123","count":10}`,
		"{\"value\":\"Test123\",\"count\":10}")

	assertPipeCmdByTestRunner(t, `drop("a")`, `{"a":1,"b":2}`, "{\"b\":2}\n")
	//assertPipeCmd(t, `drop("a")`, `{"a":1,"b":2}`, "{\"b\":2}\n")

	assertPipeCmdByTestRunner(t, `flat("a")`, `{"a":[[1], "a", [[[true]]]}`, "1\n\"a\"\ntrue\n")

	assertPipeCmdByTestRunnerError(t, `flat("")`, `{"a":[[1], "a", [[[true]]]}`,
		"field cannot be empty")

	assertPipeCmdByTestRunner(t, `grep_add("title", "(?is)test", "new_title")`,
		`{"title":"Test123"}
{"title":"123test456"}`,
		`{"title":"Test123","new_title":[["Test"]]}
{"title":"123test456","new_title":[["test"]]}`)

	// 正常
	assertPipeCmdByTestRunner(t, `rm("title")`,
		`{"title":"abc","a":1}`,
		`{"a":1}
`)

	// 字段不存在
	assertPipeCmdByTestRunner(t, `rm("title")`,
		`{"a":1}`,
		`{"a":1}
`)

	// 不提供字段
	assertPipeCmdByTestRunnerError(t, `rm()`,
		`{"a":1}`,
		`rm must has field params`)

	// 提供空字段
	assertPipeCmdByTestRunnerError(t, `rm("")`,
		`{"a":1}`,
		`path cannot be empty`)

	assertPipeCmdByTestRunner(t, `sort("a")`, `{"a":2}
{"a":1}`, `{"a":1}
{"a":2}
`)

	assertPipeCmdByTestRunner(t, `sort()`, "1\n2\n1\n", `1
1
2
`)

	assertPipeCmdByTestRunner(t, `stats("a")`, `{"a":1}
{"a":2}
{"a":1}
`, `{"value":2,"count":1}
{"value":1,"count":2}
`)

	// stats("a", 1) 等同于 value("a") | stats("", 1)
	assertPipeCmdByTestRunner(t, `stats("a", 1)`, `{"a":1}
{"a":2}
{"a":1}
`, `{"value":1,"count":2}
`)

	//
	assertPipeCmdByTestRunner(t, `stats("", 1)`, `1
2
1
`, `{"value":1,"count":2}
`)

	assertPipeCmdByTestRunner(t, `to_int("a")`, `{"a":"2"}`, `{"a":2}
`)

	assertPipeCmdByTestRunner(t, `uniq()`, "1\n2\n1\n", "1\n2\n1\n")
	assertPipeCmdByTestRunner(t, `uniq()`, "1\n1\n2\n", "1\n2\n")
	assertPipeCmdByTestRunner(t, `uniq(true)`, "1\n2\n1\n", `{"value":1,"count":1}
{"value":2,"count":1}
{"value":1,"count":1}
`)
	assertPipeCmdByTestRunner(t, `uniq(true)`, "1\n1\n2\n", `{"value":1,"count":2}
{"value":2,"count":1}
`)

	// 先sort再uniq
	assertPipeCmdByTestRunner(t, `sort() & uniq(true)`, "1\n2\n1\n", `{"value":1,"count":2}
{"value":2,"count":1}
`)

	assertPipeCmdByTestRunner(t, `value("a")`, `{"a":1}`, "1\n")

	assertPipeCmdByTestRunner(t, `zq("a")`, `{"a":1}`, "{\"a\":1}\n")
}

func TestLoad_screenshot(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wait, _ := strconv.Atoi(r.FormValue("wait"))
		if wait > 0 {
			time.Sleep(time.Second * time.Duration(wait))
		}
		w.Write([]byte("hello world"))
	}))
	defer ts.Close()

	// 截图正确
	assert.Contains(t, assertPipeCmd(t, `gen("{\"host\":\"`+ts.URL+`\"}") & screenshot("host")`, ``), "screenshot_filepath")
	assert.Contains(t, assertPipeCmd(t, `gen("{\"url\":\"`+ts.URL+`\"}") & screenshot()`, ``), "screenshot_filepath")
	assert.Contains(t, assertPipeCmd(t, `gen("{\"url\":\"`+ts.URL+`\"}") & screenshot("url","sc_filepath")`, ``), "sc_filepath")
	assert.Contains(t, assertPipeCmd(t, `gen("{\"url\":\"`+ts.URL+`\"}") & screenshot("url","sc_filepath",10)`, ``), "sc_filepath")
	// 超时
	assert.NotContains(t, assertPipeCmd(t, `gen("{\"url\":\"`+ts.URL+`?wait=10\"}") & screenshot("url","sc_filepath",1)`, ``), "sc_filepath")
	// 截图异常
	assert.NotContains(t, assertPipeCmd(t, `gen("{\"host\":\"http://127.0.0.1:55\"}") & screenshot("host")`, ``), "screenshot_filepath")
}

func TestLoad_fork(t *testing.T) {
	ast := workflowast.NewParser().MustParse(`load("./data/test.json") & [cut("a") | cut("b")]`)
	p := New()
	_, err := p.Run(ast)
	assert.Nil(t, err)
	res, err := os.ReadFile(p.LastFile)
	assert.Nil(t, err)
	assert.Equal(t, `{"a":1,"b":2}`, string(res))
	assert.Equal(t, 2, len(p.LastTask.Children))
	res, err = os.ReadFile(p.LastTask.Children[0].GetLastFile())
	assert.Nil(t, err)
	assert.Equal(t, "{\"a\":1}\n", string(res))
	res, err = os.ReadFile(p.LastTask.Children[1].GetLastFile())
	assert.Nil(t, err)
	assert.Equal(t, "{\"b\":2}\n", string(res))
}

func TestLoad_fofa(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/info/my":
			w.Write([]byte(`{"error":false,"email":"` + r.FormValue("email") + `","fcoin":10,"isvip":true,"vip_level":1}`))
		case "/api/v1/search/all":
			w.Write([]byte(`{"error":false,"size":12345678,"page":1,"mode":"extended","query":"host=\"https://fofa.info\"","results":[["fofa1.info"]]}`))
		}
	}))
	defer ts.Close()

	var err error
	var fofacli *gofofa.Client
	code := workflowast.NewParser().MustParse(`fofa("host=\"https://fofa1.info\"", "domain", 1)`)
	fofacli, err = gofofa.NewClient(gofofa.WithURL(ts.URL))
	assert.Nil(t, err)
	p := New().WithObject(gocodefuncs.FofaObjectName, fofacli)

	_, err = p.Run(code)
	assert.Nil(t, err)

	content, err := os.ReadFile(p.LastFile)
	assert.Nil(t, err)
	assert.Equal(t, `{"domain":"fofa1.info"}
`,
		string(content))
}

func assertFileContent(t *testing.T, filename string, content string) {
	data, err := os.ReadFile(filename)
	assert.Nil(t, err)
	assert.Equal(t, content, string(data))
}

func assertPipeRunnerContent(t *testing.T, p *PipeRunner, content string) {
	assertFileContent(t, p.LastFile, content)
}

func TestPipeRunner_fix_url(t *testing.T) {
	var err error
	var code string

	p := New()
	code, err = workflowast.NewParser().Parse(`gen("{\"url\":\"1.1.1.1:81\"}") & fix_url()`)
	assert.Nil(t, err)
	_, err = p.Run(code)
	assert.Nil(t, err)
	assertPipeRunnerContent(t, p, "{\"url\":\"http://1.1.1.1:81\"}\n")

	p = New()
	_, err = p.Run(workflowast.NewParser().MustParse(`gen("{\"host\":\"1.1.1.1:81\"}") & fix_url("host")`))
	assert.Nil(t, err)
	assertPipeRunnerContent(t, p, "{\"host\":\"http://1.1.1.1:81\"}\n")

	p = New()
	_, err = p.Run(workflowast.NewParser().MustParse(`gen("{\"url\":\"1.1.1.1:80\"}") & fix_url()`))
	assert.Nil(t, err)
	assertPipeRunnerContent(t, p, "{\"url\":\"http://1.1.1.1\"}\n")

	p = New()
	_, err = p.Run(workflowast.NewParser().MustParse(`gen("{\"url\":\"https://1.1.1.1:443\"}") & fix_url()`))
	assert.Nil(t, err)
	assertPipeRunnerContent(t, p, "{\"url\":\"https://1.1.1.1\"}\n")

	p = New()
	_, err = p.Run(workflowast.NewParser().MustParse(`gen("{\"host\":\"1.1.1.1:81\"}") & fix_url("")`))
	assert.Error(t, err)
}

func TestPipeRunner_DumpTasks(t *testing.T) {
	tpl, err := template.New("tasks").Funcs(template.FuncMap{
		"HasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
	}).Parse(`{{ if HasPrefix . "aaa" }}yes{{ end }}`)
	assert.Nil(t, err)
	var out bytes.Buffer
	err = tpl.Execute(&out, "aaa")
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "yes", out.String())

	p := New()
	_, err = p.Run(workflowast.NewParser().MustParse(`gen("{\"a\":\"1\",\"b\":2}") & [cut("a") | cut("b")]`))
	assert.Nil(t, err)
	c := p.DumpTasks(false, "", sync.Map{})
	assert.Contains(t, c, "fork")

	// 截图
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	}))
	defer ts.Close()

	p.Close()
	_, err = p.Run(workflowast.NewParser().MustParse(`gen("{\"host\":\"` + ts.URL + `\"}") & screenshot("host")`))
	assert.Nil(t, err)
	assert.Equal(t, "image/png", p.LastTask.Result.Artifacts[0].FileType)
	c = p.DumpTasks(true, "", sync.Map{})
	assert.Contains(t, c, "<img")

	p.Close()
	_, err = p.Run(workflowast.NewParser().MustParse(`gen("{\"host\":\"` + ts.URL + `\"}") & screenshot("host")`))
	assert.Nil(t, err)
	assert.Equal(t, "image/png", p.LastTask.Result.Artifacts[0].FileType)
	c = p.DumpTasks(true, "/abcprefix", sync.Map{})
	assert.Contains(t, c, "/abcprefix/")

}

func TestPipeRunner_Close(t *testing.T) {
	p := New()
	_, err := p.Run(workflowast.NewParser().MustParse(`load("./data/test.json") | cut("a")`))
	assert.Nil(t, err)
	c, err := os.ReadFile(p.LastFile)
	assert.Nil(t, err)
	assert.True(t, len(c) > 0)

	p.Close()
	_, err = os.ReadFile(p.LastFile)
	assert.Error(t, err)
}

func TestPipeRunner_toExcel(t *testing.T) {
	p := New()
	code := workflowast.NewParser().MustParse(`gen("{\"a\":1,\"b\":\"2\"}") & to_excel()`)
	_, err := p.Run(code)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(p.LastTask.Result.Artifacts))
	f, err := excelize.OpenFile(p.LastTask.Result.Artifacts[0].FilePath)
	assert.Nil(t, err)
	v, err := f.GetCellValue("Sheet1", "A1")
	assert.Nil(t, err)
	assert.Equal(t, "a", v)
	v, err = f.GetCellValue("Sheet1", "A2")
	assert.Nil(t, err)
	assert.Equal(t, "1", v)
	v, err = f.GetCellValue("Sheet1", "B2")
	assert.Nil(t, err)
	assert.Equal(t, "2", v)
}

func assertToSql(t *testing.T, workFlowName string, dsn string, db *sql.DB) {

	// 只生成文件
	p := New()
	code := workflowast.NewParser().MustParse(`gen("{\"a\":1,\"b\":\"2\",\"c\":true}") & ` + workFlowName + `("tbl")`)
	_, err := p.Run(code)
	assert.Nil(t, err)
	switch workFlowName {
	case "to_sqlite":
		assert.Equal(t, 2, len(p.LastTask.Result.Artifacts))
	case "to_mysql":
		assert.Equal(t, 1, len(p.LastTask.Result.Artifacts))
	default:
		panic("unknown workFlowName: " + workFlowName)
	}

	d, err := os.ReadFile(p.LastTask.Result.Artifacts[0].FilePath)
	assert.Nil(t, err)
	assert.Equal(t, `INSERT INTO tbl (a,b,c) VALUES (1,"2",true)
`, string(d))

	// 分叉测试
	p.Close()
	code = workflowast.NewParser().MustParse(`gen("{\"a\":1,\"b\":\"2\",\"c\":\"3\"}") & [flat("a") | ` + workFlowName + `("tbl","a,b")]`)
	_, err = p.Run(code)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(p.LastTask.Children))
	switch workFlowName {
	case "to_sqlite":
		assert.Equal(t, 2, len(p.LastTask.Children[1].LastTask.Result.Artifacts))
	case "to_mysql":
		assert.Equal(t, 1, len(p.LastTask.Children[1].LastTask.Result.Artifacts))
	default:
		panic("unknown workFlowName: " + workFlowName)
	}

	d, err = os.ReadFile(p.LastTask.Children[1].LastTask.Result.Artifacts[0].FilePath)
	assert.Nil(t, err)
	assert.Equal(t, `INSERT INTO tbl (a,b) VALUES (1,"2")
`, string(d))

	if db != nil {
		checkRow := func(rows *sql.Rows) {
			if rows.Next() {
				var a int
				var b string
				err = rows.Scan(&a, &b)
				assert.Nil(t, err)

				assert.Equal(t, `2`, b)
				assert.Equal(t, 1, a)
			}
		}

		var rows *sql.Rows
		// 有字段
		p.Close()
		code = workflowast.NewParser().MustParse(`gen("{\"a\":1,\"b\":\"2\",\"c\":\"3\"}") & ` + workFlowName + `("tbl","a,b","` + dsn + `")`)
		_, err = p.Run(code)
		assert.Nil(t, err)
		d, err = os.ReadFile(p.LastTask.Result.Artifacts[0].FilePath)
		assert.Nil(t, err)
		assert.Equal(t, `INSERT INTO tbl (a,b) VALUES (1,"2")
`, string(d))
		rows, err = db.Query("select a,b from tbl")
		checkRow(rows)

		// 没有字段，自动提取
		p.Close()
		params := ""
		switch workFlowName {
		case "to_sqlite":
			params = `"` + dsn + `"`
		case "to_mysql":
			params = `"","` + dsn + `"`
		default:
			panic("unknown workFlowName: " + workFlowName)
		}
		code = workflowast.NewParser().MustParse(`gen("{\"a\":1,\"b\":\"2\",\"c\":\"3\"}") & ` + workFlowName + `("tbl",` + params + `)`)
		_, err = p.Run(code)
		assert.Nil(t, err)
		d, err = os.ReadFile(p.LastTask.Result.Artifacts[0].FilePath)
		assert.Nil(t, err)
		assert.Equal(t, `INSERT INTO tbl (a,b) VALUES (1,"2")
`, string(d))
		rows, err = db.Query("select a,b from tbl")
		checkRow(rows)
	}
}

func TestPipeRunner_toSqlite(t *testing.T) {
	assertToSql(t, "to_sqlite", "", nil)
}

func TestPipeRunner_toMysql(t *testing.T) {
	var err error
	var d []byte

	var db *sql.DB
	var dsn string

	if utils.DockerStatusOk() {
		// 用docker来跑mysql进行测试
		d, err = utils.DockerRun("run", "--rm", "--detach", "--name", "gofofamysqltest", "--env", "MARIADB_ROOT_PASSWORD=my-secret-pw", "--env", "MYSQL_ROOT_PASSWORD=my-secret-pw", "-p13306:3306", "mariadb")
		assert.Nil(t, err)
		assert.NotContains(t, string(d), "Error response from daemon")
		assert.Regexp(t, "[0-9a-f]{64}", string(d))
		defer func() {
			_, err = utils.DockerRun("stop", "gofofamysqltest")
			assert.Nil(t, err)
		}()

		// 等待mariadb下载完成
		s := time.Now()
		for time.Since(s) < time.Minute {
			d, err = utils.RunCmdNoExitError(utils.DockerRun("ps"))
			if err == nil && strings.Contains(string(d), "gofofamysqltest") {
				break
			}
			time.Sleep(time.Second)
		}

		// 取IP
		d, err = utils.DockerRun("inspect", "gofofamysqltest")
		assert.Nil(t, err)
		var r *regexp.Regexp
		r = regexp.MustCompile(`"IPAddress": "(.*?)"`)
		matched := r.FindAllStringSubmatch(string(d), 1)
		assert.True(t, len(matched) > 0)
		cip := matched[0][1]
		assert.True(t, len(cip) > 0)

		// 等待启动,10s
		for i := 0; i < 10; i++ {
			d, err = utils.DockerRun("run", "--rm", "mariadb", "mysql", "-h", cip, "-uroot", "-pmy-secret-pw", "-e", "select @@version")
			if strings.Contains(string(d), "-MariaDB-") {
				break
			}
			time.Sleep(time.Second)
		}

		d, err = utils.DockerRun("run", "--rm", "mariadb", "mysql", "-h", cip, "-uroot", "-pmy-secret-pw", "-e", "create database aaa; create database bbb; use aaa; CREATE TABLE tbl ( a varchar(255), b varchar(255)); select @@version")
		assert.Nil(t, err)
		assert.Contains(t, string(d), "-MariaDB-")

		// 创建数据表测试
		dsn = "root:my-secret-pw@tcp(127.0.0.1:13306)/aaa"
		db, err = sql.Open("mysql", dsn)
		assert.Nil(t, err)
		err = db.Ping()
		assert.Nil(t, err)
		assertToSql(t, "to_mysql", dsn, db)

		// 不创建数据表测试
		dsn = "root:my-secret-pw@tcp(127.0.0.1:13306)/bbb"
		db, err = sql.Open("mysql", dsn)
		assert.Nil(t, err)
		assertToSql(t, "to_mysql", dsn, db)
	}

	assertToSql(t, "to_mysql", dsn, db)
}

func TestPipeRunner_Run(t *testing.T) {
	// callid测试
	p := New()
	ast := workflowast.NewParser()
	code := ast.MustParse("gen(`{\"port\":1,\"ip\":\"1.1.1.1\"}`) & to_int(`port`) & sort(`port`) & [cut(`port`) | cut(`ip`)]")
	_, err := p.Run(code)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(ast.CallList))
	assert.Equal(t, 5, len(p.Tasks))
	for i := range ast.CallList {
		assert.Equal(t, strconv.Itoa(ast.CallList[i].UUID), p.Tasks[i].ActionID)
	}

	p = New().WithUserFunction([]interface{}{
		"FetchFofa", gocodefuncs.GenFofaFieldData,
	})

	ast = workflowast.NewParser()
	code = ast.MustParse("fofa(\"title=test\",\"host,ip,port,country\", 1000) & [flat(\"port\") & sort() & uniq(true) & sort(\"count\") & zq(\"tail 10\") & chart(\"pie\") | flat(\"country\") & sort() & uniq(true) & sort(\"count\") & zq(\"tail 10\") & chart(\"pie\") | zq(\"tail 1\") & screenshot(\"host\") & to_excel() | to_sqlite(\"tbl\", \"host,ip,port\")]")
	_, err = p.Run(code)
	assert.Nil(t, err)

	assert.Equal(t, 17, len(ast.CallList))
	assert.Equal(t, 17, len(p.Tasks))
	for i := range ast.CallList {
		assert.Equal(t, strconv.Itoa(ast.CallList[i].UUID), p.Tasks[i].ActionID)
	}
}

func TestPipeRunner_scanPort(t *testing.T) {
	d, err := utils.ExecCmdWithTimeout(2*time.Second, "nmap", "-v")
	if err == nil && strings.Contains(string(d), "Starting Nmap") {
		// 启动服务器
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello world"))
		}))
		defer ts.Close()
		u, _ := url.Parse(ts.URL)

		// 扫描
		p := New()
		ast := workflowast.NewParser()
		code := ast.MustParse("scan_port(`" + u.Hostname() + "`,`" + u.Port() + "`)")
		_, err = p.Run(code)
		assert.Nil(t, err)

		d, err = utils.ReadFirstLineOfFile(p.LastFile)
		assert.Nil(t, err)
		assert.Contains(t, string(d), u.Port())
	}

}

func TestPipeRunner_if_add(t *testing.T) {
	p := New()
	ast := workflowast.NewParser()
	code := ast.MustParse("gen(`{\"port\":80,\"ip\":\"1.1.1.1\"}`) & if_add(`has(protocol)`,`host`,`ip+\":\"+port`)")
	_, err := p.Run(code)
	assert.Nil(t, err)
	d, err := utils.ReadFirstLineOfFile(p.LastFile)
	assert.Nil(t, err)
	assert.NotContains(t, utils.JSONLineFields(string(d)), "host")

	p.Close()
	code = ast.MustParse("gen(`{\"port\":80,\"ip\":\"1.1.1.1\",\"protocol\":\"http\"}`) & if_add(`has(protocol)`,`host`,`ip+\":\"+port`)")
	_, err = p.Run(code)
	assert.Nil(t, err)
	d, err = utils.ReadFirstLineOfFile(p.LastFile)
	assert.Nil(t, err)
	assert.Contains(t, utils.JSONLineFields(string(d)), "host")

	p.Close()
	code = ast.MustParse("gen(`{\"port\":80,\"ip\":\"1.1.1.1\",\"protocol\":\"http\"}`) & if_add(`protocol==\"http\"`,`host`,`ip+\":\"+port`)")
	_, err = p.Run(code)
	assert.Nil(t, err)
	d, err = utils.ReadFirstLineOfFile(p.LastFile)
	assert.Nil(t, err)
	assert.Contains(t, utils.JSONLineFields(string(d)), "host")

	p.Close()
	code = ast.MustParse("gen(`{\"port\":80,\"ip\":\"1.1.1.1\",\"protocol\":\"https\"}`) & if_add(`protocol==\"http\"`,`host`,`ip+\":\"+port`)")
	_, err = p.Run(code)
	assert.Nil(t, err)
	d, err = utils.ReadFirstLineOfFile(p.LastFile)
	assert.Nil(t, err)
	assert.NotContains(t, utils.JSONLineFields(string(d)), "host")

	p.Close()
	code = ast.MustParse("gen(`{\"port\":80,\"ip\":\"1.1.1.1\",\"protocol\":\"https\"}`) & if_add(`protocol==\"http\" OR protocol==\"https\"`,`host`,`ip+\":\"+port`)")
	_, err = p.Run(code)
	assert.Nil(t, err)
	d, err = utils.ReadFirstLineOfFile(p.LastFile)
	assert.Nil(t, err)
	assert.Contains(t, utils.JSONLineFields(string(d)), "host")
}

func TestPipeRunner_WithWebHook(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		d, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Write(d)
		called = true
	}))
	defer ts.Close()
	p := New().WithWebHook(ts.URL)
	_, err := p.Run(`GenFofaFieldData(GetRunner(), map[string]interface{} {
	"query": "test",
	"size": 1000,
	"fields": "host,ip,port,title",
})

FlatArray(GetRunner(), map[string]interface{}{
    "field": "port",
})

ZqQuery(GetRunner(), map[string]interface{}{
    "query": "sort",
})

ZqQuery(GetRunner(), map[string]interface{}{
    "query": "uniq -c",
})

ZqQuery(GetRunner(), map[string]interface{}{
    "query": "sort count",
})

ZqQuery(GetRunner(), map[string]interface{}{
    "query": "tail 10",
})

GenerateChart(GetRunner(), map[string]interface{}{
    "type": "pie",
})`)
	assert.True(t, called)
	assert.Nil(t, err)

	tarGzData, err := p.TarGzAll()
	assert.Nil(t, err)
	assert.True(t, len(tarGzData) > 0)

	// ungzip
	zr, err := gzip.NewReader(bytes.NewReader(tarGzData))
	assert.Nil(t, err)
	// untar
	tr := tar.NewReader(zr)

	filenum := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		assert.Nil(t, err)
		t.Log("Name: \n", header.Name)
		filenum++
	}

	err = zr.Close()
	assert.Nil(t, err)
	assert.Equal(t, 7, filenum)
}

func TestPipeRunner_TarGzAll(t *testing.T) {
	p := New()
	_, err := p.Run(`FetchFile(GetRunner(), map[string]interface{} {
	"url": "https://web.stanford.edu/class/archive/cs/cs109/cs109.1166/stuff/titanic.csv",
})`)
	assert.Nil(t, err)

	tarGzData, err := p.TarGzAll()
	assert.Nil(t, err)
	assert.True(t, len(tarGzData) > 0)

	tarGzFile, err := utils.WriteTempFile(".tar.gz", func(f *os.File) error {
		_, err = f.Write(tarGzData)
		return err
	})
	assert.Nil(t, err)

	tarGzOpen, err := os.Open(tarGzFile)
	assert.Nil(t, err)
	defer tarGzOpen.Close()

	// ungzip
	zr, err := gzip.NewReader(tarGzOpen)
	assert.Nil(t, err)
	// untar
	tr := tar.NewReader(zr)

	filenum := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		assert.Nil(t, err)
		t.Log("Name: \n", header.Name)
		filenum++
	}

	err = zr.Close()
	assert.Nil(t, err)
	assert.Equal(t, 2, filenum)
}
