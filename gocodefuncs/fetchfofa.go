package gocodefuncs

import (
	"context"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/LubyRuffy/gofofa"
	"github.com/LubyRuffy/gofofa/pkg/outformats"
	"github.com/mitchellh/mapstructure"
	"os"
	"strings"
)

// FetchFofaParams 获取fofa的参数
type FetchFofaParams struct {
	Query     string
	Size      int
	Fields    string
	Frequency float32 `json:"frequency"`
}

var (
	FofaObjectName         = "fofaCli"
	FetchMaxSizeObjectName = "fetch_max_size"
	DefaultFetchMaxSize    = 100000 // 最大获取记录数的大小
)

// FetchFofa 从fofa获取数据
func FetchFofa(p Runner, params map[string]interface{}) *FuncResult {
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

	options.Query = ExpendVarWithJsonLine(p, options.Query, "")

	maxSize, ok := p.GetObject(FetchMaxSizeObjectName)
	if !ok {
		maxSize = DefaultFetchMaxSize
	}
	if options.Size > maxSize.(int) {
		panic(fmt.Errorf("max size greater than: %d", maxSize))
	}

	fields := strings.Split(options.Fields, ",")

	var res [][]string
	fofaCli, ok := p.GetObject(FofaObjectName)
	if !ok {
		panic(fmt.Errorf("HostSearch failed: doesn't set " + FofaObjectName))
	}

	// 设置context
	ctx, cancel := context.WithCancel(p.GetContext())
	defer cancel()
	fofaCli.(*gofofa.Client).SetContext(ctx)

	res, err = fofaCli.(*gofofa.Client).HostSearch(options.Query, options.Size, fields)
	if err != nil {
		panic(fmt.Errorf("HostSearch failed: %w", err))
	}

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		w := outformats.NewJSONWriter(f, fields)
		return w.WriteAll(res)
	})
	if err != nil {
		panic(fmt.Errorf("fetchFofa error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}

func init() {
	RegisterObject(FofaObjectName, "should be gofofa.Client")
	RegisterObject(FetchMaxSizeObjectName, fmt.Sprintf("should be int, default %d", DefaultFetchMaxSize))
}
