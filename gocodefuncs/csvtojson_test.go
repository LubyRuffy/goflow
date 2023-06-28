package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"os"
	"testing"
)

func writeSampleCSVFile() string {
	filename, err := utils.WriteTempFile(".csv", func(f *os.File) error {
		_, err := f.WriteString("a,b\n1,2")
		return err
	})
	if err != nil {
		panic(err)
	}
	return filename
}

func TestCSVToJson(t *testing.T) {
	filename := writeSampleCSVFile()
	fr := CSVToJson(&testRunner{
		T:        t,
		lastFile: filename,
	}, map[string]interface{}{})
	f, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	gjson.ParseBytes(f).ForEach(func(key, value gjson.Result) bool {
		assert.Equal(t, key.String(), "Sheet1")
		assert.Equal(t, value.String(), `[["a","b"],["1","2"]]`)
		return false
	})

}
