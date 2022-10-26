package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToExcel(t *testing.T) {
	// 读excel-》json
	filename := writeSampleExcelFile()
	fr := ExcelToJson(&testRunner{
		T:        t,
		lastFile: filename,
	}, map[string]interface{}{})

	// json=》写excel
	result := ToExcel(&testRunner{
		T:        t,
		lastFile: fr.OutFile,
	}, map[string]interface{}{
		"rawFormat": true,
	})

	// 再读
	fr = ExcelToJson(&testRunner{
		T:        t,
		lastFile: result.OutFile,
	}, map[string]interface{}{})
	f, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"Sheet1":[["IP","域名"],["1.1.1.1","a.com"]],"Sheet2":[null,["Hello world."]]}`)
}
