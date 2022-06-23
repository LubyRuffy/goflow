package utils

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestFixURL(t *testing.T) {
	cases := [][]string{
		{"1.1.1.1:81", "http://1.1.1.1:81"},
		{"1.1.1.1:80", "http://1.1.1.1"},
		{"http://1.1.1.1:80", "http://1.1.1.1"},
		{"1.1.1.1:443", "https://1.1.1.1"},
		{"https://1.1.1.1:443", "https://1.1.1.1"},
		{"https://1.1.1.1:8443", "https://1.1.1.1:8443"},
		{"https://1.1.1.1:443/a?timeout=1", "https://1.1.1.1/a?timeout=1"},
		{"https://1.1.1.1:443/a?timeout=1#1", "https://1.1.1.1/a?timeout=1"},
		{"https://a:b@1.1.1.1:443/a?timeout=1#1", "https://1.1.1.1/a?timeout=1"},
	}
	for _, test := range cases {
		assert.Equal(t, test[1], FixURL(test[0]))
	}
}

func TestHttpHeaderToString(t *testing.T) {
	v := HttpHeaderToString(http.Header{
		"a": []string{"1"},
		"b": []string{"2", "3"},
	})
	// map 顺序时随机的
	assert.True(t, v == "a: 1\nb: 2,3\n" || v == "b: 2,3\na: 1\n")
}
