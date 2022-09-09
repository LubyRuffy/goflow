package gocodefuncs

import (
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_chromeActions(t *testing.T) {
	buf := []byte{}
	type args struct {
		in      chromeActionsInput
		logf    func(string, ...interface{})
		timeout int
		actions []chromedp.Action
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试正常screen picture",
			args: args{
				in: chromeActionsInput{
					URL: "https://www.baidu.com",
				},
				logf:    func(s string, i ...interface{}) {},
				timeout: 10,
				actions: []chromedp.Action{
					chromedp.Sleep(time.Second * time.Duration(1)),
					chromedp.CaptureScreenshot(&buf),
				},
			},
		},
		{
			name: "测试自定义proxy & UA",
			args: args{
				in: chromeActionsInput{
					URL:       "http://www.baidu.com",
					Proxy:     "socks5://127.0.0.1:7890",
					UserAgent: "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36",
				},
				logf:    func(s string, i ...interface{}) {},
				timeout: 10,
				actions: []chromedp.Action{
					chromedp.Sleep(time.Second * time.Duration(1)),
					chromedp.CaptureScreenshot(&buf),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := chromeActions(tt.args.in, func(s string, i ...interface{}) {}, tt.args.timeout, tt.args.actions...)
			assert.Nil(t, err)
			assert.NotNil(t, buf)
			t.Logf("screenshot result: %s", buf[0:5])
		})
	}
}
