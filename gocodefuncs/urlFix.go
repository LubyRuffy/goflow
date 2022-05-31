package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"net"
	"os"
	"strings"
)

// UrlFix 自动补齐url
func UrlFix(p Runner, params map[string]interface{}) *FuncResult {
	var fn string
	var err error
	field := "url"
	if len(params) > 0 {
		field = params["urlField"].(string)
	}
	if len(field) == 0 {
		panic(fmt.Errorf("urlFix must has a field"))
	}

	fn, err = utils.WriteTempFile("", func(f *os.File) error {
		return utils.EachLine(p.GetLastFile(), func(line string) error {
			v := gjson.Get(line, field).String()
			if !strings.Contains(v, "://") {
				host, port, _ := net.SplitHostPort(v)
				if port == "80" {
					v = host
				}
				v = "http://" + v
			}
			line, err := sjson.Set(line, field, v)
			if err != nil {
				return err
			}
			_, err = f.WriteString(line + "\n")
			return err
		})
	})
	if err != nil {
		panic(fmt.Errorf("urlFix failed: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
