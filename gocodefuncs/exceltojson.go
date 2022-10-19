package gocodefuncs

import (
	"encoding/json"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/xuri/excelize/v2"
	"io"
	"os"
)

// ExcelToJsonParams 获取fofa的参数
type ExcelToJsonParams struct {
}

func readExcel(f io.Reader) (interface{}, error) {
	excelF, err := excelize.OpenReader(f)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, sheet := range excelF.GetSheetList() {
		// Get all the rows in the Sheet1.
		rows, err := excelF.GetRows(sheet)
		if err != nil {
			return nil, err
		}

		result[sheet] = rows
	}

	return result, nil
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

	excelF, err := os.Open(p.GetLastFile())
	if err != nil {
		panic(fmt.Errorf("read excel failed: %w", err))
	}
	defer excelF.Close()

	records, err := readExcel(excelF)
	if err != nil {
		panic(fmt.Errorf("read excel failed: %w", err))
	}

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		jsonStr, err := json.Marshal(records)

		_, err = f.Write(jsonStr)
		return err
	})
	if err != nil {
		panic(fmt.Errorf("read excel error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
