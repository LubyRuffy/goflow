package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
	"testing"
)

func writeSampleExcelFile() string {
	filename := "Book1.xlsx"
	f := excelize.NewFile()
	// Create a new worksheet.
	index := f.NewSheet("Sheet2")
	// Set value of a cell.
	f.SetCellValue("Sheet2", "A2", "Hello world.")
	f.SetCellValue("Sheet1", "A1", "IP")
	f.SetCellValue("Sheet1", "B1", "域名")
	f.SetCellValue("Sheet1", "A2", "1.1.1.1")
	f.SetCellValue("Sheet1", "B2", "a.com")
	// Set the active worksheet of the workbook.
	f.SetActiveSheet(index)
	// Save the spreadsheet by the given path.
	if err := f.SaveAs(filename); err != nil {
		fmt.Println(err)
	}
	return filename
}

func TestExcelToJson(t *testing.T) {
	filename := writeSampleExcelFile()
	fr := ExcelToJson(&testRunner{
		T:        t,
		lastFile: filename,
	}, map[string]interface{}{})
	f, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"Sheet1":[["IP","域名"],["1.1.1.1","a.com"]],"Sheet2":[null,["Hello world."]]}`)

	// 单列合并单元格
	fr = ExcelToJson(&testRunner{
		T:        t,
		lastFile: "../data/a.xlsx",
	}, map[string]interface{}{})
	f, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(f), `{"Sheet1":[["t1","t2","t3"],["dct11","dct12","dct13"],["dct11","dct22","dct13"]],"Sheet2":[["t1","t2","t3"],["dct11","dct11","dct12"],["dct11","dct11","dct22"]],"_merged_Sheet1":[["A2:A3","dct11"],["C2:C3","dct13"]],"_merged_Sheet2":[["A2:B3","dct11"]]}`)

	// 多列合并单元格
}
