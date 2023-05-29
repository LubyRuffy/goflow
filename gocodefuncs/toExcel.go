package gocodefuncs

import (
	"context"
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

	// 创建空白单元格样式
	style, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	err = utils.EachLineWithContext(context.TODO(), formattedFile, func(line string) error {
		if options.RawFormat {
			v := gjson.ParseBytes([]byte(line))
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
						// 格式化单元格
						err = f.SetColWidth(key.String(), fmt.Sprintf("%c", colNo+int32(cols)),
							fmt.Sprintf("%c", colNo+int32(cols)), 30)
						if err != nil {
							panic(fmt.Errorf("ToExcel SetColWidth failed: %w", err))
						}
						err = f.SetRowHeight(key.String(), rows+1, 35)
						if err != nil {
							panic(fmt.Errorf("ToExcel SetRowHeight failed: %w", err))
						}
						err = f.SetCellStyle(key.String(), fmt.Sprintf("%c%d", colNo+int32(cols), rows+1),
							fmt.Sprintf("%c%d", colNo+int32(cols), rows+1), style)
						if err != nil {
							panic(fmt.Errorf("ToExcel SetCellStyle failed: %w", err))
						}
						// 写入内容
						err = f.SetCellValue(key.String(), fmt.Sprintf("%c%d", colNo+int32(cols), rows+1),
							value.Array()[rows].Array()[cols].Value())
						if err != nil {
							panic(fmt.Errorf("ToExcel SetCellValue failed: %w", err))
						}
					}
				}

				// 单元格合并，格式 "_merged_Sheet2":[["A2:B3","dct11"]]}
				/**
				当 key 中存在 "." 时，直接使用 v.Get 会被解析为多层嵌套 json，需要用 v.Map() 直接指定
				*/
				sheetName := key.String()
				if mergedCells, ok := v.Map()["_merged_"+sheetName]; ok && mergedCells.Exists() && mergedCells.IsArray() {
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

					// 如果配置了写入图片项，直接设置图片
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
		return nil
	})
	if err != nil {
		return nil
	}

	// auto merge 选项，检查（上下、左右）相邻的多个格子，如果内容一致则进行合并操作
	autoMergeExcel(f)

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

func autoMergeExcel(f *excelize.File) {
	/**
	首先按行检查，按行合并后再通过列合并
	*/
	// 获取所有的工作表名称
	sheetNames := f.GetSheetMap()

	// 遍历每个工作表
	for _, sheetName := range sheetNames {
		// 读取工作表内容
		rows, err := f.GetRows(sheetName)
		if err != nil {
			fmt.Println(err)
			return
		}

		colNo := 'A'
		startIndex := 0
		// 按行合并 todo: 按列合并
		for i, row := range rows {
			startContent := rows[i][startIndex]
			for j, cell := range row {
				if j == 0 {
					continue
				}
				if startContent == cell {
					err = f.MergeCell(sheetName, fmt.Sprintf("%c%d", colNo+int32(j), i+1),
						fmt.Sprintf("%c%d", colNo+int32(j), i+1))
					if err != nil {
						panic(fmt.Errorf("MergeCell failed: %w", err))
					}
				} else {
					startIndex = j
				}
			}
		}
	}
}
