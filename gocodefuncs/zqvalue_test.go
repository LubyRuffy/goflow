package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestZqValue(t *testing.T) {
	filename, err := utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"ts":"2022-12-17 09:00:00"}
{"ts":"2022-12-18 09:00:00"}`)
		return err
	})
	assert.Nil(t, err)
	v := ZqValue(&testRunner{
		T:        t,
		lastFile: filename,
	}, map[string]interface{}{
		"query": `yield time(this.ts) | v:=max(this) | yield string(this.v) | yield replace(this, "Z", "") | yield replace(this, "T", " ")`,
	})
	assert.Equal(t, `2022-12-18 09:00:00`, v)
}
