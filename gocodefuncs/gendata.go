package gocodefuncs

import (
	"encoding/json"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"os"
)

// GenData 生成数据
func GenData(p Runner, params map[string]interface{}) *FuncResult {
	var fn string
	var err error

	ext := ".txt"
	var m map[string]interface{}
	if err = json.Unmarshal([]byte(params["data"].(string)), &m); err == nil {
		ext = ".json"
	}
	fn, err = utils.WriteTempFile(ext, func(f *os.File) error {
		_, err = f.WriteString(params["data"].(string))
		return err
	})
	if err != nil {
		panic(fmt.Errorf("genData failed: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
