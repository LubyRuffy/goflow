package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	}))
	defer ts.Close()

	data := `{"url":"` + ts.URL + `"}`
	fr := HttpRequest(newTestRunner(data), map[string]interface{}{
		"keepBody": true,
	})
	assert.NotEqual(t, "", fr.OutFile)
	assert.Equal(t, 0, len(fr.Artifacts))
	d, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, "hello world", gjson.GetBytes(d, "url_requested.http_body").String())
	assert.Equal(t, int64(200), gjson.GetBytes(d, "url_requested.http_status").Int())
	assert.Contains(t, gjson.GetBytes(d, "url_requested.http_header").String(), "Content-Type: text/plain; charset=utf-8\n")

	//t.Log(string(d))
}
