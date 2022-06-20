package utils

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// FixURL 补充完整url，主要用于ip:port变成url
func FixURL(v string) string {
	if !strings.Contains(v, "://") {
		host, port, _ := net.SplitHostPort(v)
		switch port {
		case "80":
			v = "http://" + host
		case "443":
			v = "https://" + host
		default:
			v = "http://" + v
		}
	} else {
		u, err := url.Parse(v)
		if err != nil {
			return v
		}
		//v = u.String() 不会过滤标准端口
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

		v += u.Path
		if len(u.RawQuery) > 0 {
			v += "?" + u.RawQuery
		}
	}
	return v
}

// HttpHeaderToString http header 转换为字符串
func HttpHeaderToString(header http.Header) string {
	var r string
	for k, v := range header {
		r += k + ": " + strings.Join(v, ",") + "\n"
	}
	return r
}
