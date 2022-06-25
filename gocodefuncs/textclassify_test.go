package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"testing"
)

func TestTextClassify(t *testing.T) {
	data := `{"name":"这是要给银行测试"}
{"name":"这又是要给银行测试"}
{"name":"这是要给能源测试"}
{"name":"这是要给运营商测试"}
{"name":"这就是一个测试"}
{"name":"这是生产系统"}`
	fr := TextClassify(newTestRunner(t, data), map[string]interface{}{
		"textField": "name",
		"saveField": "name_tag",
		"filters": [][]string{
			{"银行", "银行"},
			{"运营商", "运营商"},
			{"测试", "测试"},
		},
	})

	assert.FileExists(t, fr.OutFile)
	i := 0
	utils.EachLine(fr.OutFile, func(line string) error {
		tag := gjson.Get(line, "name_tag").String()

		switch i {
		case 0, 1:
			assert.Equal(t, "银行", tag)
		case 2, 4:
			assert.Equal(t, "测试", tag)
		case 3:
			assert.Equal(t, "运营商", tag)
		case 5:
			assert.Equal(t, "", tag)
		}
		i++
		return nil
	})
}
