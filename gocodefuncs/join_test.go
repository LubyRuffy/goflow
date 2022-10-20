package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func joinQuery(c1, c2 string, t *testing.T) string {
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
		"file": fn1,
	})
	f, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	return string(f)
}

func TestJoinQuery(t *testing.T) {

	assert.Equal(t, joinQuery(`{"a":1}`, `{"b":2}`, t), `{"a":"1","b":"2"}`)
	// 冲突
	assert.Equal(t, joinQuery(`{"a":1}`, `{"a":2}`, t), `{"a":"2"}`)
	// 多行
	assert.Equal(t, joinQuery(`{"a":1}
{"c":3}`, `{"b":2}
{"d":4}`, t), `{"a":"1","c":"3","b":"2","d":"4"}`)

}
