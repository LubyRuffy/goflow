package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func writeSampleJsonFile(t *testing.T, content string) string {
	fn, err := utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(content)
		return err
	})
	assert.Nil(t, err)
	return fn
}

func testJq(t *testing.T, stream bool, content string, query string, except string) {
	filename := writeSampleJsonFile(t, content)
	fr := JqQuery(&testRunner{
		T:        t,
		lastFile: filename,
	}, map[string]interface{}{
		"query":  query,
		"stream": stream,
	})
	f, err := os.ReadFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, except, string(f))
}

func TestJqQuery(t *testing.T) {
	testJq(t, false, `{"a":[1,2], "b":[2,3]}`, ".a-.b", `[1]
`)

	testJq(t, true, `{"ip":"1.1.1.1","a":1}
{"ip":"1.1.1.1","b":2}
{"ip":"2.2.2.2","c":3}`,
		`group_by(.ip)| map({ ip: (.[0].ip) } + ([.[]|del(.ip)] | reduce .[] as $item({}; .+$item)) ) | .[]`,
		`{"a":1,"b":2,"ip":"1.1.1.1"}
{"c":3,"ip":"2.2.2.2"}
`)

	// 报错
	//testJq(t, false, `{"a":[1,2], "b":[2,3]}`, ".a-.b aaaa", `[1]`)
}
