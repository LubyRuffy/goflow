package utils

import (
	"bufio"
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"hash/fnv"
	"os/exec"
	"sort"
	"strings"
	"time"
)

func read(r *bufio.Reader) ([]byte, error) {
	var (
		isPrefix = true
		err      error
		line, ln []byte
	)

	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}

	return ln, err
}

// EscapeString 双引号内的字符串转换
func EscapeString(s string) string {
	//s, _ = sjson.Set(`{"a":""}`, "a", s)
	//return s[strings.Index(s, `:`)+1 : len(s)-1]
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// EscapeDoubleQuoteStringOfHTML 双引号内的字符串转换为Mermaid格式（HTML）
func EscapeDoubleQuoteStringOfHTML(s string) string {
	s = strings.ReplaceAll(s, `"`, `#quot;`)
	return s
}

// JSONLineFieldsWithType 获取json行的fields，包含属性信息
func JSONLineFieldsWithType(line string) (fields [][]string) {
	v := gjson.Parse(line)
	v.ForEach(func(key, value gjson.Result) bool {
		typeStr := "text"
		switch value.Type {
		case gjson.True, gjson.False:
			typeStr = "boolean"
		case gjson.Number:
			typeStr = "int"
		}
		fields = append(fields, []string{key.String(), typeStr})
		return true
	})
	return
}

// JSONLineFields 获取json行的fields
func JSONLineFields(line string) (fields []string) {
	fs := JSONLineFieldsWithType(line)
	for _, f := range fs {
		fields = append(fields, f[0])
	}
	return
}

//// GetCurrentProcessFileDir 获得当前程序所在的目录
//func GetCurrentProcessFileDir() string {
//	return filepath.Dir(os.Args[0])
//}
//
//// UserHomeDir 获得当前用户的主目录
//func UserHomeDir() string {
//	if runtime.GOOS == "windows" {
//		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
//		if home == "" {
//			home = os.Getenv("USERPROFILE")
//		}
//		return home
//	}
//	return os.Getenv("HOME")
//}

// ExecCmdWithTimeout 在时间范围内执行系统命令，并且将输出返回（stdout和stderr）
func ExecCmdWithTimeout(timeout time.Duration, arg ...string) (b []byte, err error) {
	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	routeCmd := exec.CommandContext(ctx, arg[0], arg[1:]...)

	return routeCmd.CombinedOutput()
}

// RunCmdNoExitError 将exec.ExitError不作为错误，通常配合exec.Command使用
func RunCmdNoExitError(d []byte, err error) ([]byte, error) {
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			err = nil
		}
	}
	return d, err
}

// SimpleHash hashes using fnv32a algorithm
func SimpleHash(text string) string {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(text))
	return fmt.Sprintf("0x%08x", algorithm.Sum32())
}

// MapPair map[string]int64排序后的元素对
type MapPair struct {
	Name  string
	Value int64
}
type PairList []MapPair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// TopMapByValue 根据map[string]int64的值进行排序，取topSize的条数出来
func TopMapByValue(m map[string]int64, topSize int) PairList {
	mSize := len(m)
	if topSize < 1 {
		topSize = mSize
	} else {
		if topSize > mSize {
			topSize = mSize
		}
	}
	pl := make(PairList, len(m))
	i := 0
	for k, v := range m {
		pl[i] = MapPair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl[0:topSize]
}
