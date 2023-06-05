package gocodefuncs

import (
	"context"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/xuri/excelize/v2"
	"log"
	"strings"
)

type toExcelParam struct {
	RawFormat  bool `mapstructure:"rawFormat"`  // 是否原始格式 {"Sheet1":[[]], "Sheet2":[[]]}
	InsertPic  bool `mapstructure:"insertPic"`  // 是否将截图字段自动替换为图片, rawFormat不受该参数影响
	JsonFormat bool `mapstructure:"jsonFormat"` // 从 json 直接格式化为 excel
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

	lineNum := 0
	err = utils.EachLineWithContext(context.TODO(), formattedFile, func(line string) error {
		lineNum++
		if options.JsonFormat {
			// 临时切换
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
		} else if options.RawFormat {
			// 临时修改
			err = jsonFormatToExcel(f, line, lineNum)
			if err != nil {
				return err
			}
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

	// todo: auto merge 选项，检查（上下、左右）相邻的多个格子，如果内容一致则进行合并操作
	//autoMergeExcel(f)

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

func formatWriteCell(f *excelize.File, sheetName string, row, cols int, value gjson.Result) (err error) {
	colNo := 'A'

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

	// 格式化单元格
	err = f.SetColWidth(sheetName, fmt.Sprintf("%c", colNo+int32(cols)),
		fmt.Sprintf("%c", colNo+int32(cols)), 30)
	if err != nil {
		panic(fmt.Errorf("ToExcel SetColWidth failed: %w", err))
	}
	err = f.SetRowHeight(sheetName, row+1, 35)
	if err != nil {
		panic(fmt.Errorf("ToExcel SetRowHeight failed: %w", err))
	}
	err = f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", colNo+int32(cols), row+1),
		fmt.Sprintf("%c%d", colNo+int32(cols), row+1), style)
	if err != nil {
		panic(fmt.Errorf("ToExcel SetCellStyle failed: %w", err))
	}
	// 写入内容
	err = f.SetCellValue(sheetName, fmt.Sprintf("%c%d", colNo+int32(cols), row+1),
		value.String())
	if err != nil {
		return fmt.Errorf("ToExcel SetCellValue failed: %w", err)
	}

	return nil
}

func jsonFormatToExcel(f *excelize.File, line string, lineNum int) (err error) {
	currentRow := 0
	v := gjson.ParseBytes([]byte(line))
	sheetName := fmt.Sprintf("Sheet%d", lineNum)
	index := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// 开始遍历 键值 对应关系
	v.ForEach(func(key, value gjson.Result) bool {
		/**
		列表形式：
		 "c_title": [
		        {
		            "count": 27,
		            "name": "DPTECH ONLINE",
		            "source": "fofa.info/stats"
		        },
		        {
		            "count": 12,
		            "name": "HTTP状态 404 - 未找到",
		            "source": "fofa.info/stats"
		        }
		    ]
		|       |count|         name         | source          |
		|c_title| 27  |     DPTECH ONLINE    | fofa.info/stats |
		|		| 12  |HTTP状态 404 - 未找到   | fofa.info/stats |
		*/
		if value.IsArray() {
			cols := 0
			startRow := currentRow
			// 写最左侧的 key, 稍后合并
			err = formatWriteCell(f, sheetName, startRow, cols, key)
			if err != nil {
				log.Printf("write object header cell failed %s", err.Error())
				return false
			}

			// 处理 value with key
			value.ForEach(func(k, v gjson.Result) bool {
				cols = 0
				if v.IsObject() {
					// object 按照上述输出
					v.ForEach(func(innerKey, innerValue gjson.Result) bool {
						if currentRow == startRow {
							// 第一行写对应的 key & value
							cols++
							err = formatWriteCell(f, sheetName, currentRow, cols, innerKey)
							if err != nil {
								return false
							}
							err = formatWriteCell(f, sheetName, currentRow+1, cols, innerValue)
							if err != nil {
								return false
							}
							return true
						} else {
							cols++
							// 后面的行只写对应的 value
							err = formatWriteCell(f, sheetName, currentRow, cols, innerValue)
							if err != nil {
								return false
							}
							return true
						}
					})
					if startRow == currentRow {
						currentRow += 2
					} else {
						currentRow += 1
					}
					return true
				} else {
					// 非 object，直接输出
					err = formatWriteCell(f, sheetName, currentRow, cols, k)
					if err != nil {
						return false
					}
					err = formatWriteCell(f, sheetName, currentRow, cols+1, v)
					if err != nil {
						return false
					}
					currentRow += 1
					return true
				}
			})

			// value 处理完成，合并最左边的标题
			err = f.MergeCell(sheetName, fmt.Sprintf("A%d", startRow+1), fmt.Sprintf("A%d", currentRow))
			if err != nil {
				log.Printf("merge cell failed %s", err.Error())
				return false
			}
		} else if value.IsObject() {
			/**
			object 形式：
			 "location": {
			        "city": "Hangzhou City",
			        "country": "China",
			        ...
			    }
			|        |       city     |    country   |      source     |
			|location| Hangzhou City  |     China    | fofa.info/stats |
			*/
			cols := 0
			// 写最左侧的 key, 并与下一行合并
			err = formatWriteCell(f, sheetName, currentRow, cols, key)
			if err != nil {
				log.Printf("write object header cell failed %s", err.Error())
				return false
			}
			err = f.MergeCell(sheetName, fmt.Sprintf("A%d", currentRow+1), fmt.Sprintf("A%d", currentRow+2))
			if err != nil {
				log.Printf("merge cell failed %s", err.Error())
				return false
			}
			// 写入右边的 object 对应的 key/value
			value.ForEach(func(k, v gjson.Result) bool {
				cols++
				err = formatWriteCell(f, sheetName, currentRow, cols, k)
				if err != nil {
					return false
				}
				err = formatWriteCell(f, sheetName, currentRow+1, cols, v)
				if err != nil {
					return false
				}
				return true
			})
			currentRow += 2
		} else {
			cols := 0
			/**
			flat 正常形式：
			"ip": "122.224.163.198"
			|	ip	 |  122.224.163.198  |
			*/
			// 写最左侧的 key, 并与下一行合并
			err = formatWriteCell(f, sheetName, currentRow, cols, key)
			if err != nil {
				log.Printf("write object header cell failed %s", err.Error())
				return false
			}
			err = formatWriteCell(f, sheetName, currentRow, cols+1, value)
			if err != nil {
				log.Printf("write object value cell failed %s", err.Error())
				return false
			}
			currentRow++
		}

		return true
	})

	return nil
}
