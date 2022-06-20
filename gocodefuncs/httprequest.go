package gocodefuncs

import (
	"crypto/tls"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/gammazero/workerpool"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io/ioutil"
	"net/http"
	"os"
)

// HttpRequestParams http请求的参数
type HttpRequestParams struct {
	URLField  string `json:"urlField"`  // url的字段名称，默认是url
	UserAgent string `json:"userAgent"` // 模拟的客户端，默认是defaultUserAgent
	TLSVerify bool   `json:"tlsVerify"` // 是否验证tls
	Workers   int    `json:"workers"`   // 并发限制
	MaxSize   int    `json:"maxSize"`   // 最大长度，默认是100000，需要无限制改成-1
}

// HttpRequest http请求提取数据
func HttpRequest(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options HttpRequestParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("HttpRequest failed: %w", err))
	}

	if options.URLField == "" {
		options.URLField = "url"
	}
	if options.UserAgent == "" {
		options.UserAgent = defaultUserAgent
	}
	if options.Workers == 0 {
		options.Workers = 5
	}
	if options.MaxSize == 0 {
		options.MaxSize = 100000 // 100k
	}

	// 配置是否验证tls
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !options.TLSVerify},
	}

	wp := workerpool.New(options.Workers)
	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		err = utils.EachLine(p.GetLastFile(), func(line string) error {
			wp.Submit(func() {
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
				resp, err = client.Get(url)
				if err != nil {
					p.Warnf("HttpRequest failed: %s, %s", url, err)
					f.WriteString(line + "\n")
					return
				}
				defer resp.Body.Close()

				// 不管是否成功都先把数据写入
				var body []byte
				body, err = ioutil.ReadAll(resp.Body)
				if options.MaxSize > 0 && len(body) > options.MaxSize {
					body = body[:options.MaxSize]
				}
				line, _ = sjson.Set(line, "http_body", body)
				line, _ = sjson.Set(line, "http_status", resp.StatusCode)
				line, _ = sjson.Set(line, "http_header", utils.HttpHeaderToString(resp.Header))

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
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("HttpRequest error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
