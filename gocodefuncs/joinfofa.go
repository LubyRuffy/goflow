package gocodefuncs

import (
	"context"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/LubyRuffy/gofofa"
	"github.com/avast/retry-go"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/sjson"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

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

	// fofa连接
	fofaCli, ok := p.GetObject(FofaObjectName)
	if !ok {
		panic(fmt.Errorf("HostSearch failed: doesn't set fofacli"))
	}
	// 设置context
	ctx, cancel := context.WithCancel(p.GetContext())
	defer cancel()
	fofaCli.(*gofofa.Client).SetContext(ctx)

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		err = utils.EachLineWithContext(p.GetContext(), p.GetLastFile(), func(line string) error {
			defer func() {
				atomic.AddInt64(&processed, 1)
				p.SetProgress(float64(processed) / float64(lines))
			}()

			query := ExpendVarWithJsonLine(p, options.Query, line)
			if len(query) == 0 {
				// 不用查询
				return nil
			}

			// 请求fofa
			var res [][]string

			err := retry.Do(
				func() error {
					res, err = fofaCli.(*gofofa.Client).HostSearch(query, options.Size, fields)
					if err != nil {
						return fmt.Errorf("HostSearch failed: %w", err)
					}

					return nil
				},
				retry.DelayType(retry.BackOffDelay),
				retry.Delay(time.Second*1),
				retry.MaxDelay(time.Second*3),
				retry.Attempts(3),
				retry.OnRetry(func(n uint, err error) {
					fmt.Printf("Retry %d: %s\n", n, err.Error())
				}),
			)

			if len(res) > 0 {
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
			} else {
				// fofa没有数据，就把原始内容写入
				_, err = f.WriteString(line + "\n")
				if err != nil {
					panic(err)
				}
			}

			if options.Frequency > 0 {
				time.Sleep(time.Duration(options.Frequency) * time.Second)
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
