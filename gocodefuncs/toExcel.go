package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/xuri/excelize/v2"
	"strings"
)

type toExcelParam struct {
	RawFormat bool // 是否原始格式 {"Sheet1":[[]], "Sheet2":[[]]}
}

// ToExcel 写excel文件
func ToExcel(p Runner, params map[string]interface{}) *FuncResult {
	var fn string
	var err error

	var options toExcelParam
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("screenShot failed: %w", err))
	}

	fn, err = utils.WriteTempFile(".xlsx", nil)
	if err != nil {
		panic(fmt.Errorf("toExcel failed: %w", err))
	}

	f := excelize.NewFile()
	defer f.Close()

	if options.RawFormat {
		var line []byte
		line, err = utils.ReadFirstLineOfFile(p.GetLastFile())
		if err != nil {
			panic(fmt.Errorf("ToExcel SetCellValue failed: %w", err))
		}
		v := gjson.ParseBytes(line)
		colNo := 'A'
		v.ForEach(func(key, value gjson.Result) bool {
			// 合并的记录数据
			if strings.HasPrefix(key.String(), "_merged_") {
				return true
			}

			index := f.NewSheet(key.String())
			f.SetActiveSheet(index)
			for rows := range value.Array() {
				for cols := range value.Array()[rows].Array() {
					err = f.SetCellValue(key.String(), fmt.Sprintf("%c%d", colNo+int32(cols), rows+1),
						value.Array()[rows].Array()[cols].Value())
					if err != nil {
						panic(fmt.Errorf("ToExcel SetCellValue failed: %w", err))
					}
				}
			}

			// 炒作单元格合并，格式 "_merged_Sheet2":[["A2:B3","dct11"]]}
			if mergedCells := v.Get("_merged_" + key.String()); mergedCells.Exists() && mergedCells.IsArray() {
				for _, c := range mergedCells.Array() {
					err = f.MergeCell(key.String(), c.Array()[0].String(), c.Array()[1].String())
					if err != nil {
						panic(fmt.Errorf("MergeCell failed: %w", err))
					}
				}
			}
			return true
		})
	} else {
		lineNo := 2
		err = utils.EachLineWithContext(p.GetContext(), p.GetLastFile(), func(line string) error {
			v := gjson.Parse(line)
			colNo := 'A'
			v.ForEach(func(key, value gjson.Result) bool {
				// 设置第一行
				if lineNo == 2 {
					err = f.SetCellValue("Sheet1", fmt.Sprintf("%c%d", colNo, lineNo-1), key.Value())
				}

				// 写值
				err = f.SetCellValue("Sheet1", fmt.Sprintf("%c%d", colNo, lineNo), value.Value())
				colNo++
				if err != nil {
					panic(fmt.Errorf("SetCellValue failed: %w", err))
				}
				return true
			})
			lineNo++
			return err
		})
		if err != nil {
			panic(fmt.Errorf("toExcel failed: %w", err))
		}
	}

	err = f.SaveAs(fn)
	if err != nil {
		panic(fmt.Errorf("toExcel failed: %w", err))
	}

	AddStaticResource(p, fn)
	return &FuncResult{
		//OutFile: fn,
		Artifacts: []*Artifact{
			{
				FilePath: fn,
				FileType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			},
		},
	}
}
