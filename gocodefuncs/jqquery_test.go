package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func writeSampleJsonFile(t *testing.T) string {
	fn, err := utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"a":[1,2], "b":[2,3]}`)
		return err
	})
	assert.Nil(t, err)
	return fn
}

func TestJqQuery(t *testing.T) {
	filename := writeSampleJsonFile(t)
	fr := JqQuery(&testRunner{
		T:        t,
		lastFile: filename,
	}, map[string]interface{}{
		"query": ".a-.b",
	})
	f, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `[1]`)
}
