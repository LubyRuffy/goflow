package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestJoinQuery(t *testing.T) {
	fn1, err := utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"a":1}`)
		return err
	})
	assert.Nil(t, err)
	fn2, err := utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"b":2}`)
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
	assert.Equal(t, string(f), `{"a":"1","b":"2"}`)

	// 冲突
	fn1, err = utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"a":1}`)
		return err
	})
	assert.Nil(t, err)
	fn2, err = utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"a":2}`)
		return err
	})
	assert.Nil(t, err)
	fr = Join(&testRunner{
		T:        t,
		lastFile: fn2,
	}, map[string]interface{}{
		"file": fn1,
	})
	f, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"a":"2"}`)

	// 多行
	fn1, err = utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"a":1}
{"c":3}`)
		return err
	})
	assert.Nil(t, err)
	fn2, err = utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"b":2}
{"d":4}`)
		return err
	})
	assert.Nil(t, err)
	fr = Join(&testRunner{
		T:        t,
		lastFile: fn2,
	}, map[string]interface{}{
		"file": fn1,
	})
	f, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"a":"1","c":"3","b":"2","d":"4"}`)

}
