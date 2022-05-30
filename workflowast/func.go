package workflowast

import (
	"fmt"
	"strconv"
	"sync"
)

// FunctionTranslateHook 转换hook函数
type FunctionTranslateHook func(fi *FuncInfo) string

var (
	globalFunctionTranslateHooks sync.Map
)

type FuncParameter struct {
	v interface{}
}

// Int64 做为int64返回
func (fp FuncParameter) Int64() int64 {
	return fp.v.(int64)
}

// String 做为string返回
func (fp FuncParameter) String() string {
	return fp.v.(string)
}

// Bool 做为bool返回
//func (fp FuncParameter) Bool() bool {
//	return fp.v.(bool)
//}

func rawString(s string) string {
	l := len(s)
	if l == 0 {
		return s
	}
	switch s[0] {
	case '`':
		return s[1 : l-1]
	case '"':
		// 暂时不处理转义
		return s[1 : l-1]
	}
	return s
}

// RawString 不要引号
func (fp FuncParameter) RawString() string {
	return rawString(fp.v.(string))
}

// ToString 转换成字符串
func (fp FuncParameter) ToString() string {
	switch fp.v.(type) {
	case string:
		return fp.v.(string)
	case int64:
		return strconv.FormatInt(fp.v.(int64), 10)
	case *FuncInfo:
		return fp.v.(*FuncInfo).String()
	case bool:
		if fp.v.(bool) {
			return "true"
		} else {
			return "false"
		}
	default:
		panic(fmt.Errorf("FuncParameter.ToString failed, unknown type: %v", fp.v))
	}
}

// FuncInfo 函数信息
type FuncInfo struct {
	Name   string           // 函数名称
	Params []*FuncParameter // 参数列表
	UUID   int              // 唯一的ID
}

// String func id string
func (f *FuncInfo) String() string {

	if v, ok := globalFunctionTranslateHooks.Load(f.Name); ok {
		return v.(FunctionTranslateHook)(f)
	}

	rStr := f.Name + "("
	for i, p := range f.Params {
		if i != 0 {
			rStr += ", "
		}
		rStr += p.ToString()
	}
	return rStr + ")"
}

// RegisterFunction 注册函数，包括函数名称和Hook
// hook用于生成底层的函数部分
func RegisterFunction(name string, f FunctionTranslateHook) {
	globalFunctionTranslateHooks.Store(name, f)
}
