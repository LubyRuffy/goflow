package gocodefuncs

import (
	"encoding/json"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/xuri/excelize/v2"
	"io"
	"os"
	"regexp"
)

var (
	cellRowColRegex = regexp.MustCompile(`(\s+)(\d+)`)
)

// ExcelToJsonParams 获取fofa的参数
type ExcelToJsonParams struct {
}

// A2
//func cellRC(cellAxis string) (row int, col int, err error) {
//	sRC := cellRowColRegex.FindStringSubmatch(cellAxis)
//	col, err = excelize.ColumnNameToNumber()
//
//	row, err = strconv.Atoi(sRC[1])
//	return
//}

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

		if v, err := excelF.GetMergeCells(sheet); err == nil {
			if len(v) > 0 {
				result["_merged_"+sheet] = v
				// 填充数据
				for index := range v {
					//"A2:A3"
					scs, sr, err := excelize.SplitCellName(v[index].GetStartAxis())
					if err != nil {
						return nil, err
					}
					sc, err := excelize.ColumnNameToNumber(scs)
					if err != nil {
						return nil, err
					}
					ecs, er, err := excelize.SplitCellName(v[index].GetEndAxis())
					if err != nil {
						return nil, err
					}
					ec, err := excelize.ColumnNameToNumber(ecs)
					if err != nil {
						return nil, err
					}

					sr = sr - 1
					er = er - 1
					sc = sc - 1
					ec = ec - 1
					// sc和ec应该一样？单列；多列会怎么样？
					for i := sr; i <= er; i++ {
						for j := sc; j <= ec; j++ {
							if i == sr && j == sc {
								continue
							}
							//log.Println("change value of cell ", i, j, "to", rows[sr][sc])
							if j >= len(rows[i]) {
								rows[i] = append(rows[i], rows[sr][sc])
							} else {
								rows[i][j] = rows[sr][sc]
							}

						}
					}
				}
			}

		} else {
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
