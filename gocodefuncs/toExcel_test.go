package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"os"
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
		lastFile: result.Artifacts[0].FilePath,
	}, map[string]interface{}{})
	f, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"Sheet1":[["IP","域名"],["1.1.1.1","a.com"]],"Sheet2":[null,["Hello world."]]}`)

	// json=》写excel
	json := `{"Sheet1":[["IP"],["1.1.1.1"],["2.2.2.2"]]}`
	fn, err := utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(json)
		return err
	})
	result = ToExcel(&testRunner{
		T:        t,
		lastFile: fn,
	}, map[string]interface{}{
		"rawFormat": true,
	})
	// 再读
	fr = ExcelToJson(&testRunner{
		T:        t,
		lastFile: result.Artifacts[0].FilePath,
	}, map[string]interface{}{})
	f, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"Sheet1":[["IP"],["1.1.1.1"],["2.2.2.2"]]}`)

	// 写合并的表格
	json = `{"Sheet1":[["t1","t2","t3"],["dct11","dct12","dct13"],["dct11","dct22","dct13"]],"Sheet2":[["t1","t2","t3"],["dct11","dct11","dct12"],["dct11","dct11","dct22"]],"_merged_Sheet1":[["A2:A3","dct11"],["C2:C3","dct13"]],"_merged_Sheet2":[["A2:B3","dct11"]]}`
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(json)
		return err
	})
	result = ToExcel(&testRunner{
		T:        t,
		lastFile: fn,
	}, map[string]interface{}{
		"rawFormat": true,
	})
	// 再读
	fr = ExcelToJson(&testRunner{
		T:        t,
		lastFile: result.Artifacts[0].FilePath,
	}, map[string]interface{}{})
	f, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"Sheet1":[["t1","t2","t3"],["dct11","dct12","dct13"],["dct11","dct22","dct13"]],"Sheet2":[["t1","t2","t3"],["dct11","dct11","dct12"],["dct11","dct11","dct22"]],"_merged_Sheet1":[["A2:A3","dct11"],["C2:C3","dct13"]],"_merged_Sheet2":[["A2:B3","dct11"]]}`)
}
