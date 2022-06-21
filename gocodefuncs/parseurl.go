package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/weppos/publicsuffix-go/publicsuffix"
	"net"
	"net/url"
	"os"
	"path"
	"strings"
)

type parseURLParams struct {
	URLField string
}

// ParseURL 解析url字段
func ParseURL(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options parseURLParams
	if err = mapstructure.Decode(params, &options); err != nil {
		if err != nil {
			panic(fmt.Errorf("ParseURL error: %w", err))
		}
	}

	if len(options.URLField) == 0 {
		options.URLField = "url"
		//panic(errors.New("ParseURL failed: no url field found:" + options.URLField))
	}

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		err = utils.EachLine(p.GetLastFile(), func(line string) error {
			v := gjson.Get(line, options.URLField)
			if !v.Exists() {
				_, err = f.WriteString(line + "\n")
				return err
			}

			// 存在字段
			var u *url.URL
			u, err = url.Parse(utils.FixURL(v.String()))
			if err != nil {
				return err
			}

			port := u.Port()
			if port == "" {
				switch u.Scheme {
				case "ftp":
					port = "21"
				case "ssh":
					port = "22"
				case "http":
					port = "80"
				case "https":
					port = "443"
				}
			}
			fields := map[string]interface{}{
				"url":      u,
				"host":     u.Host,
				"hostName": u.Hostname(),
				"port":     port,
				"scheme":   u.Scheme,
				"path":     u.Path,
				"dir":      path.Dir(u.Path),
				"file":     path.Base(u.Path),
				"ext":      path.Ext(u.Path),
			}
			if ip := net.ParseIP(u.Hostname()); ip == nil {
				// domain
				var d *publicsuffix.DomainName
				d, err = publicsuffix.Parse(u.Hostname())
				fields["domain"] = d.SLD + "." + d.TLD
				fields["subdomain"] = d.TRD
				if ips, err := net.LookupIP(u.Hostname()); err == nil {
					var ipStrings []string
					for i := range ips {
						ipStrings = append(ipStrings, ips[i].String())
					}
					fields["ip"] = strings.Join(ipStrings, ",")
				}
			} else {
				fields["ip"] = u.Hostname()
			}
			line, err = sjson.Set(line, options.URLField+"_parsed", fields)
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
