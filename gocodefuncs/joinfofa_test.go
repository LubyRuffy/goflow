package gocodefuncs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/LubyRuffy/gofofa"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestJoinFofa(t *testing.T) {
	assert.Equal(t, "abc.com", ExpendVarWithJsonLine(nil, "${{domain}}", `{"domain":"abc.com"}`))
	assert.Equal(t, "", ExpendVarWithJsonLine(nil, "${{domain}}", `{"domain1":"abc.com"}`))
	assert.Equal(t, "abc.com", ExpendVarWithJsonLine(nil, "${{a.domain}}", `{"a":{"domain":"abc.com"}}`))
}

func TestJoinFofa1(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/info/my":
			w.Write([]byte(`{"error":false,"email":"` + r.FormValue("email") + `","fcoin":10,"isvip":true,"vip_level":1}`))
		case "/api/v1/search/all":
			w.Write([]byte(`{"error":false,"size":12345678,"page":1,"mode":"extended","query":"host=\"https://fofa.info\"","results":[["123.1.1.1", "443"]]}`))
		}
	}))
	defer ts.Close()
	fofaCli, _ := gofofa.NewClient(gofofa.WithURL(ts.URL))

	type args struct {
		p       Runner
		params  map[string]interface{}
		runs    int
		maxSize int
	}
	tests := []struct {
		name      string
		args      args
		want      string
		errorFunc assert.ErrorAssertionFunc
	}{
		{
			name: "测试正常请求",
			args: args{
				p: newTestRunner(t, `{"note":"only work once"}`),
				params: map[string]interface{}{
					"query":     `port="443"`,
					"size":      10,
					"fields":    "ip,port",
					"frequency": 0,
				},
				runs: 1,
			},
			want:      "443",
			errorFunc: assert.NoError,
		},
		{
			name: "测试请求频率超限",
			args: args{
				p: newTestRunner(t, `{"note":"10 runs"}`),
				params: map[string]interface{}{
					"query":     `port="443"`,
					"size":      10,
					"fields":    "ip,port",
					"frequency": 0,
				},
				runs: 30,
			},
			want:      "443",
			errorFunc: assert.NoError,
		},
		{
			name: "测试请求数量超限",
			args: args{
				p: newTestRunner(t, `{}`),
				params: map[string]interface{}{
					"query":  `host="fofa.info"`,
					"size":   50000,
					"fields": "ip,port",
				},
				maxSize: 500,
				runs:    1,
			},
			want:      "",
			errorFunc: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					tt.errorFunc(t, r.(error))
				}
			}()

			start := time.Now()
			for i := 0; i < tt.args.runs; i++ {
				tt.args.p.SetObject(FofaObjectName, fofaCli)
				if tt.args.maxSize > 0 {
					tt.args.p.SetObject(FetchMaxSizeObjectName, tt.args.maxSize)
				}
				res := JoinFofa(tt.args.p, tt.args.params)
				fileBytes, err := ReadFirstLineOfFile(res.OutFile)
				assert.Nil(t, err)
				resMap := map[string]interface{}{}
				err = json.NewDecoder(bytes.NewReader(fileBytes)).Decode(&resMap)
				fmt.Println(fmt.Sprintf("%d : %s", i, fileBytes))
				assert.Nil(t, err)
				assert.Equal(t, tt.want, resMap["port"])
			}
			fmt.Println(time.Now().Sub(start))
		})
	}
}

// ReadFirstLineOfFile 读取文件的第一行
func ReadFirstLineOfFile(fn string) ([]byte, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var b [1]byte
	var data []byte
	for {
		_, err = f.Read(b[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return data, err
		}
		if b[0] == '\n' {
			break
		}
		data = append(data, b[0])
	}
	return data, nil
}
