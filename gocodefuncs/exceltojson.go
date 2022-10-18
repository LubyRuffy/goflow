package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/sjson"
	"github.com/xuri/excelize/v2"
	"os"
)

// ExcelToJsonParams 获取fofa的参数
type ExcelToJsonParams struct {
}

// ExcelToJson 从excel读取内容到json
func ExcelToJson(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options ExcelToJsonParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("fetchFofa failed: %w", err))
	}

	if p.GetLastFile() == "" {
		panic("no file to read")
	}

	excelF, err := excelize.OpenFile(p.GetLastFile())
	if err != nil {
		panic(fmt.Errorf("read excel failed: %w", err))
	}
	defer func() {
		// Close the spreadsheet.
		if err = excelF.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		line := ""
		for _, sheet := range excelF.GetSheetList() {
			// Get all the rows in the Sheet1.
			rows, err := excelF.GetRows(sheet)
			if err != nil {
				return err
			}

			line, err = sjson.Set(line, sheet, rows)
			if err != nil {
				return err
			}
		}

		_, err = f.WriteString(line)
		return err
	})
	if err != nil {
		panic(fmt.Errorf("read excel error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
