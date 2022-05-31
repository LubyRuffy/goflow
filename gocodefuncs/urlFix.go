package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"net"
	"net/url"
	"os"
	"strings"
)

func fixURL(v string) string {
	if !strings.Contains(v, "://") {
		host, port, _ := net.SplitHostPort(v)
		if port == "80" {
			v = host
		}
		v = "http://" + v
	} else {
		u, err := url.Parse(v)
		if err != nil {
			return v
		}
		v = u.Scheme + "://" + u.Hostname()
		var defaultPort bool
		switch u.Scheme {
		case "http":
			if u.Port() == "80" {
				defaultPort = true
			}
		case "https":
			if u.Port() == "443" {
				defaultPort = true
			}
		}
		if !defaultPort {
			v += ":" + u.Port()
		}
	}
	return v
}

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
			if len(v) == 0 {
				// 没有字段，直接写回原始行
				_, err = f.WriteString(line + "\n")
				return err
			}

			line, err = sjson.Set(line, field, fixURL(v))
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
