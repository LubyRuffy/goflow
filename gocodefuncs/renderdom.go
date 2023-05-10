package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/gammazero/workerpool"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/net/context"
)

// renderURLDOM 生成单个url的domhtml
func renderURLDOM(p Runner, in chromeActionsInput, timeout int, options *ScreenshotParam) (string, error) {
	p.Debugf("render url dom: %s", in.URL)

	var actions []chromedp.Action
	if options.Sleep > 0 {
		actions = append(actions, chromedp.Sleep(time.Second*time.Duration(options.Sleep)))
	}

	var title string
	actions = append(actions, chromedp.Title(&title))
	var url string
	actions = append(actions, chromedp.Location(&url))

	var html string
	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.GetDocument().Do(ctx)
		if err != nil {
			return err
		}
		html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		return err
	}))

	err := chromeActions(in, p.Warnf, timeout, actions...)
	if err != nil {
		return "", fmt.Errorf("renderURLDOM failed(%w): %s", err, in.URL)
	}

	return html, err
}

// RenderDOM 动态渲染指定的URL，拼凑HTML
func RenderDOM(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options ScreenshotParam
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("screenShot failed: %w", err))
	}

	if options.URLField == "" {
		options.URLField = "url"
	}
	if options.SaveField == "" {
		options.SaveField = "rendered_html"
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
					return
				}
				if !strings.Contains(u, "://") {
					u = "http://" + u
				}

				var html string
				html, err = renderURLDOM(p, chromeActionsInput{
					URL:       u,
					Proxy:     options.Proxy,
					UserAgent: options.UserAgent,
				}, options.Timeout, &options)

				// 不管是否成功都先把数据写入
				line, err = sjson.Set(line, options.SaveField, html)
				if err != nil {
					return
				}
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
		panic(fmt.Errorf("screenShot error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
