package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
	"regexp"
)

type textClassifyParam struct {
	TextField string     // 来源字段
	SaveField string     // 保存字段
	Filters   [][]string // 过滤器，第一个为分类标签，第二个为正则表达式
}

// TextClassify 文本分类，根据正则进行标签的输出
func TextClassify(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options textClassifyParam
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("TextClassify failed: %w", err))
	}

	if options.TextField == "" {
		panic(fmt.Errorf("TextClassify failed: no textField"))
	}

	// 编译所有的regexp
	if len(options.Filters) == 0 {
		panic(fmt.Errorf("TextClassify failed: no filters"))
	}

	type regItem struct {
		Tag    string
		Regexp *regexp.Regexp
	}
	var regexFilterList []regItem
	for i := range options.Filters {
		r, err := regexp.Compile(options.Filters[i][0])
		if err != nil {
			panic(fmt.Errorf("TextClassify failed: regex is not valid: %s", options.Filters[i][1]))
		}
		regexFilterList = append(regexFilterList, regItem{
			Tag:    options.Filters[i][1],
			Regexp: r,
		})
	}

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		err = utils.EachLineWithContext(p.GetContext(), p.GetLastFile(), func(line string) error {
			v := gjson.Get(line, options.TextField)
			if !v.Exists() {
				_, err = f.WriteString(line + "\n")
				return err
			}

			tag := ""
			for i := range regexFilterList {
				if regexFilterList[i].Regexp.MatchString(v.String()) {
					tag = regexFilterList[i].Tag
					break
				}
			}

			// 判断是本地文件还是远程文件还是base64
			line, err = sjson.Set(line, options.SaveField, tag)
			if err != nil {
				return err
			}
			_, err = f.WriteString(line + "\n")
			return err

		})
		return err
	})

	return &FuncResult{
		OutFile: fn,
	}
}

/*
- 只输出一个tag？还是可以多个tag？
	目前只输出一个tag就停止
- 是否又顺序要求？
	目前考虑顺序要求，前面的先匹配

	TextClassify(GetRunner(), map[string]interface{} {
		"field": "text",
		"saveField": "phishing",
		"filters": [][]string{
			{"工商银行", ".*?工商.*?"},
		},
	})
*/
