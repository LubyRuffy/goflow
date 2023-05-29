package gocodefuncs

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_renderURLDOM(t *testing.T) {
	type args struct {
		in      chromeActionsInput
		timeout int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "测试chatgpt dom渲染 1",
			args: args{
				in: chromeActionsInput{
					URL: "https://chat.jdd966.com/#/chat/1002",
				},
				timeout: 30,
			},
			want:    "未经授权",
			wantErr: assert.NoError,
		},
		{
			name: "测试chatgpt dom渲染 2",
			args: args{
				in: chromeActionsInput{
					URL: "https://chat.novapps.com/#/chat/1002",
				},
				timeout: 30,
			},
			want:    "钉钉登录",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestRunner(t, "")
			got, err := renderURLDOM(p, tt.args.in, tt.args.timeout, &ScreenshotParam{Sleep: 10})
			if !tt.wantErr(t, err, fmt.Sprintf("renderURLDOM(%v, %v, %v)", p, tt.args.in, tt.args.timeout)) {
				return
			}
			assert.Containsf(t, got, tt.want, "renderURLDOM(%v, %v, %v)", p, tt.args.in, tt.args.timeout)
		})
	}
}
