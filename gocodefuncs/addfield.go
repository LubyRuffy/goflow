package gocodefuncs

import (
	"errors"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
	"regexp"
	"strings"
)

type addFieldFrom struct {
	Method  string `json:"method"`
	Field   string
	Value   string
	Options string
}

type addFieldParams struct {
	Name  string
	Value *string       // 可以没有，就取from
	From  *addFieldFrom // 可以没有，就取Value
}

// AddField 新增字段
func AddField(p Runner, params map[string]interface{}) *FuncResult {

	var err error
	var options addFieldParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(err)
	}

	if options.Value == nil && options.From == nil {
		panic(fmt.Errorf("addField failed: neithor value nor from"))
	}

	var addRegex *regexp.Regexp

	var newLines []string
	err = utils.EachLineWithContext(p.GetContext(), p.GetLastFile(), func(line string) error {
		var newLine string
		if options.Value != nil {
			addValue := ExpendVarWithJsonLine(p, *options.Value, line)
			if len(addValue) == 0 {
				// 不用查询
				return nil
			}
			newLine, _ = sjson.Set(line, options.Name, addValue)
		} else {
			switch options.From.Method {
			case "grep":
				if addRegex == nil {
					addRegex, err = regexp.Compile(options.From.Value)
					if err != nil {
						return err
					}
				}
				res := addRegex.FindAllStringSubmatch(gjson.Get(line, options.From.Field).String(), -1)
				newLine, err = sjson.Set(line, options.Name, res)
				if err != nil {
					return err
				}
			default:
				panic(errors.New("unknown from type"))
			}
		}
		newLines = append(newLines, newLine)
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("addField error: %w", err))
	}

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		content := strings.Join(newLines, "\n")
		_, err := f.WriteString(content)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("addField error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
