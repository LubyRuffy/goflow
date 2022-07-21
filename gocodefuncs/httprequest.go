package gocodefuncs

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/gammazero/workerpool"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

// HttpRequestParams http请求的参数
type HttpRequestParams struct {
	URLField  string `json:"urlField"`  // url的字段名称，默认是url
	UserAgent string `json:"userAgent"` // 模拟的客户端，默认是defaultUserAgent
	TLSVerify bool   `json:"tlsVerify"` // 是否验证tls
	KeepBody  bool   `json:"keepBody"`  // 是否保存body
	Workers   int    `json:"workers"`   // 并发限制
	MaxSize   int    `json:"maxSize"`   // 最大长度，默认是100000，需要无限制改成-1
	TimeOut   int    `json:"timeOut"`   // 等待超时，单位为s，默认10s
	Method    string `json:"method"`    // http请求method，默认是GET
	Data      string `json:"data"`      // http请求的正文，默认为空
}

// HttpRequest http请求提取数据
func HttpRequest(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options HttpRequestParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("HttpRequest failed: %w", err))
	}

	if options.URLField == "" {
		panic(fmt.Errorf("urlField can not be empty"))
	}
	if options.UserAgent == "" {
		options.UserAgent = defaultUserAgent
	}
	if options.Method == "" {
		options.Method = http.MethodGet
	}
	if options.Workers == 0 {
		options.Workers = 5
	}
	if options.MaxSize == 0 {
		options.MaxSize = 100000 // 100k
	}
	if options.TimeOut == 0 {
		options.TimeOut = 10
	}

	// 配置是否验证tls
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !options.TLSVerify},
	}
	// 设置超时
	timeout := time.Second * time.Duration(options.TimeOut)
	tr.ResponseHeaderTimeout = timeout
	tr.IdleConnTimeout = timeout
	tr.TLSHandshakeTimeout = timeout
	tr.ExpectContinueTimeout = timeout

	var lines int64
	if lines, err = utils.FileLines(p.GetLastFile()); err != nil {
		panic(fmt.Errorf("ParseURL error: %w", err))
	}
	if lines == 0 {
		return &FuncResult{}
	}

	var processed int64

	wp := workerpool.New(options.Workers)
	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		var wpErr error
		err = utils.EachLineWithContext(p.GetContext(), p.GetLastFile(), func(line string) error {
			wp.Submit(func() {
				defer func() {
					atomic.AddInt64(&processed, 1)
					p.SetProgress(float64(processed) / float64(lines))
				}()

				select {
				case <-p.GetContext().Done():
					wpErr = context.Canceled
					return
				default:
				}

				u := gjson.Get(line, options.URLField).String()
				if len(u) == 0 {
					// 没有字段，直接写回原始行
					_, err = f.WriteString(line + "\n")
					if err != nil {
						panic(err)
					}
				}

				url := utils.FixURL(u)
				var resp *http.Response
				client := &http.Client{Transport: tr}

				var dataReader io.Reader
				if len(options.Data) > 0 {
					dataReader = bytes.NewReader([]byte(options.Data))
				}

				req, err := http.NewRequest(options.Method, url, dataReader)
				if err != nil {
					p.Warnf("HttpRequest failed: %s, %s", url, err)
					f.WriteString(line + "\n")
					return
				}
				resp, err = client.Do(req)
				if err != nil {
					p.Warnf("HttpRequest failed: %s, %s", url, err)
					f.WriteString(line + "\n")
					return
				}

				fields := map[string]interface{}{
					"http_status": resp.StatusCode,
					"http_header": utils.HttpHeaderToString(resp.Header),
				}

				if options.KeepBody {
					defer resp.Body.Close()

					// 不管是否成功都先把数据写入
					var body []byte
					body, err = ioutil.ReadAll(resp.Body)
					if options.MaxSize > 0 && len(body) > options.MaxSize {
						body = body[:options.MaxSize]
					}
					fields["http_body"] = string(body)
				}

				line, err = sjson.Set(line, options.URLField+"_requested", fields)

				_, err = f.WriteString(line + "\n")
				if err != nil {
					return
				}
			})
			return nil
		})
		if err != nil {
			return err
		}
		wp.StopWait()
		return wpErr
	})
	if err != nil {
		panic(fmt.Errorf("HttpRequest error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
