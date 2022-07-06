package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/LubyRuffy/gofofa"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	varReg = regexp.MustCompile(`${{(.*?)}}`)
)

// expandStringWithVars展开变量，第一个参数是带有变量的query，第二个参数是json字符串，第三个参数是变量替换对象（replaceText->varFile)
func expandStringWithVars(query string, jsonLine string, replaceMap sync.Map) string {
	replaceMap.Range(func(key, value interface{}) bool {
		v := gjson.Get(jsonLine, value.(string))
		if !v.Exists() {
			// 字段不存在，就不能执行查询了
			query = ""
			return false
		}

		query = strings.ReplaceAll(query, key.(string), v.String())
		return true
	})
	return query
}

// JoinFofa 根据json行从fofa获取数据并且展开
func JoinFofa(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options FetchFofaParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("fetchFofa failed: %w", err))
	}

	if len(options.Query) == 0 {
		panic(fmt.Errorf("fofa query cannot be empty"))
	}
	if len(options.Fields) == 0 {
		panic(fmt.Errorf("fofa fields cannot be empty"))
	}

	fields := strings.Split(options.Fields, ",")

	var lines int64
	if lines, err = utils.FileLines(p.GetLastFile()); err != nil {
		panic(fmt.Errorf("ParseURL error: %w", err))
	}
	if lines == 0 {
		return &FuncResult{}
	}
	var processed int64

	// 解析变量
	var vars sync.Map
	ms := varReg.FindAllStringSubmatch(options.Query, -1)
	for i := range ms {
		vars.Store(ms[i][0], ms[i][1]) // domain="${{parsed_domain}}" => ${{parsed_domain}}, parsed_domain
	}

	// fofa连接
	fofaCli, ok := p.GetObject(FofaObjectName)
	if !ok {
		panic(fmt.Errorf("HostSearch failed: doesn't set fofacli"))
	}

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		err = utils.EachLineWithContext(p.GetContext(), p.GetLastFile(), func(line string) error {
			defer func() {
				atomic.AddInt64(&processed, 1)
				p.SetProgress(float64(processed) / float64(lines))
			}()

			query := expandStringWithVars(options.Query, line, vars)
			if len(query) == 0 {
				// 不用查询
				return nil
			}

			// 请求fofa
			var res [][]string
			res, err = fofaCli.(*gofofa.Client).HostSearch(query, options.Size, fields)
			if err != nil {
				panic(fmt.Errorf("HostSearch failed: %w", err))
			}

			for i := range res {
				newLine := line
				for j := range fields {
					newLine, _ = sjson.Set(newLine, fields[j], res[i][j])
				}
				_, err = f.WriteString(newLine + "\n")
				if err != nil {
					panic(err)
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("HttpRequest error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}

}

/*
支持变量替换，变量的名称对应上一步中的字段

    // fofa获取并扩展数据， 根据domain查询，根据domain进行聚合
    JoinFofa(GetRunner(), map[string]interface{} {
		"query": "type=subdomain && domain=\"${{domain}}\"",
		"size": 10,
		"fields": "host,domain,fid",
	})

*/
