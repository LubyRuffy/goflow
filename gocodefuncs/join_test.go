package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func joinQuery(c1, c2 string, field string, t *testing.T) string {
	fn1, err := utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(c1)
		return err
	})
	assert.Nil(t, err)
	fn2, err := utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(c2)
		return err
	})
	assert.Nil(t, err)
	fr := Join(&testRunner{
		T:        t,
		lastFile: fn2,
	}, map[string]interface{}{
		"file":  fn1,
		"field": field,
	})
	f, err := os.ReadFile(fr.OutFile)
	//utils.ReadFirstLineOfFile()
	assert.Nil(t, err)
	return string(f)
}

func TestJoinQuery(t *testing.T) {

	assert.Equal(t, joinQuery(`{"a":1}`, `{"b":2}`, "", t), `{"a":1,"b":2}`)
	// 冲突
	assert.Equal(t, joinQuery(`{"a":1}`, `{"a":2}`, "", t), `{"a":2}`)
	// 多行
	assert.Equal(t, joinQuery(`{"a":1}
{"c":3}`, `{"b":2}
{"d":4}`, "", t), `{"a":1,"c":3,"b":2,"d":4}`)

	// 数组
	assert.Equal(t, joinQuery(`{"a":[1,2]}`, `{"b":["3","4"]}`, "", t), `{"a":[1,2],"b":["3","4"]}`)

	// 带field
	assert.Equal(t, joinQuery(`{"ip":"1.1.1.1","a":1}
{"ip":"1.1.1.1","b":2}
{"ip":"2.2.2.2","c":3}`, ``, "ip", t), `{"a":1,"b":2,"ip":"1.1.1.1"}
{"c":3,"ip":"2.2.2.2"}
`)
}

//func TestJoinQueryWithField(t *testing.T) {
//
//	assert.Equal(t, joinQuery(`{"a":1}
//{"a":2}`, `{"a":1,"b":2}`, "a", t), `{"a":"1","b":"2"}
//{"a":2}`)
//
//}
