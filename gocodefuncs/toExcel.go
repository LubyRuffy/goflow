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
	RawFormat bool `mapstructure:"rawFormat"` // 是否原始格式 {"Sheet1":[[]], "Sheet2":[[]]}
	InsertPic bool `mapstructure:"insertPic"` // 是否将截图字段自动替换为图片, rawFormat不受该参数影响
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

	var formattedFile string
	if options.InsertPic {
		formattedFile = p.GetLastFile()
	} else {
		// 格式化资源字段
		formattedFile, err = p.FormatResourceFieldInJson(p.GetLastFile())
		if err != nil {
			panic(fmt.Errorf("format resource field in json failed: %w", err))
		}
	}

	if options.RawFormat {
		var line []byte
		line, err = utils.ReadFirstLineOfFile(formattedFile)
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
		err = utils.EachLineWithContext(p.GetContext(), formattedFile, func(line string) error {
			v := gjson.Parse(line)
			colNo := 'A'
			v.ForEach(func(key, value gjson.Result) bool {
				// 设置第一行
				if lineNo == 2 {
					err = f.SetCellValue("Sheet1", fmt.Sprintf("%c%d", colNo, lineNo-1), key.Value())
				}

				// 设置图片
				if flds, ok := p.GetObject(utils.ResourceFieldsObjectName); ok && options.InsertPic {
					// 逐行进行文件名替换
					var file string
					for _, fld := range flds.([]string) {
						if fld == key.String() {
							file = gjson.Get(line, key.String()).String()
							err = f.AddPicture("Sheet1", fmt.Sprintf("%c%d", colNo, lineNo), file,
								`{"autofit": true}`)
							if err != nil {
								return false
							}
							// 完成，这个框里不写文字了
							return true
						}
					}
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
