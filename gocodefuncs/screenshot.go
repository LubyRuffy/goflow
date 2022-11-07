package gocodefuncs

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/LubyRuffy/goflow/utils"
	"github.com/chromedp/chromedp"
	"github.com/gammazero/workerpool"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/net/context"
)

var (
	defaultUserAgent   = "goflow/1.0"
	GlobalProxy        = "proxy"
	GlobalUserAgent    = "userAgent"
	UseGlobalProxy     = "useProxy"
	UseGlobalUserAgent = "useGlobalUserAgent"
)

type ScreenshotParam struct {
	URLField  string `json:"urlField"`            // url的字段名称，默认是url
	Timeout   int    `json:"timeout"`             // 整个浏览器操作超时
	Workers   int    `json:"workers"`             // 并发限制
	SaveField string `json:"saveField"`           // 保存截图地址的字段
	Sleep     int    `json:"sleep"`               // 加载完成后的等待事件，默认doc加载完成就截图
	Proxy     string `json:"proxy,omitempty"`     // 用户自定义代理，为空时不配置
	UserAgent string `json:"userAgent,omitempty"` // 用户自定义UA，为空时不配置
	AddUrl    bool   `json:"addUrl"`              // 在截图中展示url地址
}

type chromeActionsInput struct {
	URL       string `json:"url"`
	Proxy     string `json:"proxy,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`
}

//chromeActions 完成chrome的headless操作
func chromeActions(in chromeActionsInput, logf func(string, ...interface{}), timeout int, actions ...chromedp.Action) error {
	var err error

	// set user-agent
	if in.UserAgent == "" {
		in.UserAgent = defaultUserAgent
	}

	// prepare the chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("incognito", true), // 隐身模式
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.IgnoreCertErrors,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.UserAgent(in.UserAgent), // chromedp.Flag("user-agent", defaultUserAgent)
		chromedp.WindowSize(1024, 768),
	)

	// set proxy if exists
	if in.Proxy != "" {
		opts = append(opts, chromedp.ProxyServer(in.Proxy))
	}

	allocCtx, bcancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer bcancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(logf))
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	realActions := []chromedp.Action{
		chromedp.ActionFunc(func(cxt context.Context) error {
			// 等待完成，要么是body出来了，要么是资源加载完成
			ch := make(chan error, 1)
			go func(eCh chan error) {
				err := chromedp.Navigate(in.URL).Do(cxt)
				if err != nil {
					eCh <- err
				}
				var htmlDom string
				err = chromedp.WaitReady("body", chromedp.ByQuery).Do(cxt)
				if err == nil {
					if err := chromedp.OuterHTML("html", &htmlDom).Do(cxt); err != nil {
						log.Println("[DEBUG] fetch html failed:", err)
					}
				}
				// 20211219发现如果存在JS前端框架 (如vue, react...) 执行等待读取.
				html2Low := strings.ToLower(htmlDom)
				if strings.Contains(html2Low, "javascript") || strings.Contains(html2Low, "</script>'") {
					err = chromedp.WaitVisible("div", chromedp.ByQuery).Do(cxt)
					if err := chromedp.OuterHTML("html", &htmlDom).Do(cxt); err != nil {
						log.Println("[DEBUG] fetch html failed:", err)
					}
				}

				eCh <- err
			}(ch)

			select {
			case <-time.After(time.Duration(timeout) * time.Second):
			case err := <-ch:
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}

	realActions = append(realActions, actions...)

	// run task list
	err = chromedp.Run(ctx, realActions...)

	return err
}

func screenshotURL(p Runner, u string, options *ScreenshotParam) (string, int, error) {
	p.Debugf("screenshot url: %s", u)

	var buf []byte
	var actions []chromedp.Action
	if options.Sleep > 0 {
		actions = append(actions, chromedp.Sleep(time.Second*time.Duration(options.Sleep)))
	}
	actions = append(actions, chromedp.CaptureScreenshot(&buf))

	err := chromeActions(chromeActionsInput{
		URL:       u,
		Proxy:     options.Proxy,
		UserAgent: options.UserAgent,
	}, p.Debugf, options.Timeout, actions...)
	if err != nil {
		return "", 0, fmt.Errorf("screenShot failed(%w): %s", err, u)
	}

	if options.AddUrl == true {
		tmp, err := AddUrlToTitle(u, buf)
		if err != nil {
			log.Printf("add title failed for(%s): %s", u, err.Error())
		} else {
			buf = tmp
		}
	}

	var fn string
	fn, err = utils.WriteTempFile(".png", func(f *os.File) error {
		_, err = f.Write(buf)
		return err
	})

	return fn, len(buf), err
}

// Screenshot 截图
func Screenshot(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options ScreenshotParam
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("screenShot failed: %w", err))
	}

	if options.URLField == "" {
		options.URLField = "url"
	}
	if options.SaveField == "" {
		options.SaveField = "screenshot_filepath"
	}
	if options.Timeout == 0 {
		options.Timeout = 30
	}
	if options.Workers == 0 {
		options.Workers = 5
	}

	// 配置代理：积木块Proxy > 全局 proxy
	if UseGlobalValue(p, UseGlobalProxy) {
		options.Proxy = GetRuntimeValue(p, GlobalProxy, options.Proxy)
	}

	// 配置自定义UA：积木块 > 全局
	if UseGlobalValue(p, UseGlobalUserAgent) {
		options.UserAgent = GetRuntimeValue(p, GlobalUserAgent, options.UserAgent)
	}

	var artifacts []*Artifact

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

				var size int
				var sfn string
				url := utils.FixURL(u)
				sfn, size, err = screenshotURL(p, url, &options)
				if err != nil {
					p.Warnf("screenshotURL failed: %s, %s", url, err)
					f.WriteString(line + "\n")
					return
				}

				// 不管是否成功都先把数据写入
				line, err = sjson.Set(line, options.SaveField, sfn)
				if err != nil {
					return
				}
				_, err = f.WriteString(line + "\n")
				if err != nil {
					return
				}

				artifacts = append(artifacts, &Artifact{
					FilePath: sfn,
					FileSize: size,
					FileType: "image/png",
					FileName: filepath.Base(sfn),
					Memo:     u,
				})
				AddResource(p, sfn)
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
		panic(fmt.Errorf("screenShot error: %w", err))
	}

	AddResourceField(p, options.SaveField)

	return &FuncResult{
		OutFile:   fn,
		Artifacts: artifacts,
	}
}

// GetRuntimeValue 在 defaultValue 为空时，获取Runner中的环境变量并返回
func GetRuntimeValue(p Runner, name, defaultValue string) string {
	if defaultValue == "" {
		if value, ok := p.GetObject(name); ok {
			return value.(string)
		}
	}
	return defaultValue
}

// UseGlobalValue 根据存储的key决定是否使用全局变量
func UseGlobalValue(p Runner, name string) bool {
	if value, ok := p.GetObject(name); ok {
		if use, ok := value.(bool); ok && use {
			return true
		}
	}
	return false
}

//AddResourceField 在object中添加资源字段
func AddResourceField(p Runner, field string) {
	AddObjectSlice(p, utils.ResourceFieldsObjectName, field)
}

//AddResource 在object中添加资源列表
func AddResource(p Runner, resource string) {
	AddObjectSlice(p, utils.ResourcesObjectName, resource)
}

//AddStaticResource 在object中添加static资源
func AddStaticResource(p Runner, resource string) {
	AddObjectSlice(p, utils.StaticResourceObjectName, resource)
}

//AddObjectSlice 在object
func AddObjectSlice(p Runner, objectName, ele string) {
	var result []string
	if res, ok := p.GetObject(objectName); ok {
		if result, ok = res.([]string); !ok {
			result = []string{}
		}
	} else {
		result = []string{}
	}
	result = append(result, ele)
	p.SetObject(objectName, result)
}

//AddUrlToTitle 通过html转换对整个screenshot截图结果进行处理，添加标题栏并在其中写入访问的url地址
func AddUrlToTitle(url string, picBuf []byte) (result []byte, err error) {
	htmlPart1 := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Title</title>
    <style>
        .window {
                border-radius: 5px;
                -moz-box-shadow:1em 1em 3em #333333; -webkit-box-shadow:1em 1em 3em #333333; box-shadow:1em 1em 3em #333333;
                margin: 25px;
            }
        .window-header .btn {
            width: 10px;
            height: 10px;
            margin: 6px 0 6px 10px;
            border-radius: 50%;
            padding: 0;
            display: inline-block;
            font-size: 14px;
            font-weight: 400;
            line-height: 1.42857143;
            text-align: center;
            white-space: nowrap;
            vertical-align: middle;
            touch-action: manipulation;
            cursor: pointer;
        }
        .window-header .btn.red {
            border: 1px solid #ff3125;
            background-color: #ff6158;
        }
        .window-header .btn.yellow {
            border: 1px solid #f9ab00;
            background-color: #ffbd2d;
        }
        .window-header .btn.green {
            border: 1px solid #21a435;
            background-color: #2ace43;
        }
        .window-header {
            display: block;
            border-radius: 5px 5px 0 0;
            border-top: solid 1px #f3f1f3;
            background-image: -webkit-linear-gradient(#e3dfe3,#d0cdd0);
            background-image: linear-gradient(#e3dfe3,#d0cdd0);
            width: 100%;
            height: 22px;
        }
        body {
            font-family: "Helvetica Neue",Helvetica, "microsoft yahei", arial, STHeiTi, sans-serif;
        }
    </style>
</head>
<body>
    <div>
        <div class="window">
            <div class="window-header">
                <div class="btn red"></div>
                <div class="btn yellow"></div>
                <div class="btn green"></div>
                <div class="btn" style="margin-top: -7px;margin-left: 1%;">
                    <b style="color:#48576a">`
	htmlPart2 := `</b>
                </div>
            </div>
            <div style="max-height:800px;overflow:hidden;">
                <img  style="width:100%;" src="data:image/png;base64,`
	htmlPart3 := `" />
            </div>
        </div>
    </div>
</body>
</html>`

	// 生成的图片通过base64加密
	encodedBase64 := base64.StdEncoding.EncodeToString(picBuf)

	// 合成新的html文件
	html := append([]byte(htmlPart1), []byte(url)...)
	html = append(append(append(html, []byte(htmlPart2)...), []byte(encodedBase64)...), []byte(htmlPart3)...)
	var fn string
	fn, err = utils.WriteTempFile(".html", func(f *os.File) error {
		_, err = f.Write(html)
		return err
	})

	// 将html文件进行截图
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	var buf []byte
	if err = chromedp.Run(ctx, fullScreenshot(`file://`+fn, 90, &buf)); err != nil {
		return nil, err
	}

	return buf, err
}

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Note: chromedp.FullScreenshot overrides the device's emulation settings. Use
// device.Reset to reset the emulation and viewport settings.
func fullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.FullScreenshot(res, quality),
	}
}
