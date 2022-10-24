package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
)

type joinParams struct {
	File  string
	Field string // 合并的字段，可以为空
}

// Join 合并文件字段，一个文件是{"a":1}，另一个文件时{"b":1},则合并所有的字段为{"a":1,"b":1}
// 有冲突（相同字段）应该如何处理？默认覆盖
// 是否支持多行，还是只处理一行？支持多行，全部合并
func Join(p Runner, params map[string]interface{}) *FuncResult {

	var err error
	var options joinParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(err)
	}

	var f1Data, f2Data []byte
	if options.File != "" {
		f1Data, err = os.ReadFile(options.File)
		if err != nil {
			panic(err)
		}
	}

	if p.GetLastFile() != "" {
		f2Data, err = os.ReadFile(p.GetLastFile())
		if err != nil {
			panic(err)
		}
	}

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		if options.Field == "" {
			joinFunc := func(data []byte, line string) string {
				if data != nil {
					j := gjson.ParseBytes(data) // 可以处理多行，每行一个json没有问题
					j.ForEach(func(key, value gjson.Result) bool {
						line, err = sjson.Set(line, key.String(), value.Value())
						if err != nil {
							panic(fmt.Errorf("join error: %w", err))
						}
						return true
					})
				}
				return line
			}

			line := ""
			line = joinFunc(f1Data, line)
			line = joinFunc(f2Data, line)
			_, err = f.WriteString(line)
		} else {
			tempFile, err := utils.WriteTempFile(".json", func(f *os.File) error {
				_, err := f.Write(f1Data)
				_, err = f.WriteString("\n")
				_, err = f.Write(f2Data)
				return err
			})
			if err != nil {
				panic(err)
			}

			inFile, err := os.Open(tempFile)
			if err != nil {
				panic(err)
			}
			defer inFile.Close()

			replaceField := `"` + options.Field + `"`
			err = doJqQuery(inFile, jqQueryParams{
				Query:  fmt.Sprintf(`group_by(.%s) | map({ %s: (.[0].%s) } + ([.[]|del(.%s)] | reduce .[] as $item({}; .+$item)) ) | .[]`, replaceField, replaceField, replaceField, replaceField),
				Stream: true,
			}, func(bytes []byte) {
				_, err = f.Write(bytes)
			})
		}

		return err
	})

	return &FuncResult{
		OutFile:   fn,
		Artifacts: nil,
	}
}
