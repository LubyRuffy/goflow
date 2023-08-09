package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"testing"
)

func TestParseURL(t *testing.T) {
	data := `{"url":"http://www.baidu.com/a/b/c.html?id=1"}`
	fr := ParseURL(newTestRunner(t, data), map[string]interface{}{})
	assert.NotEqual(t, "", fr.OutFile)
	assert.Equal(t, 0, len(fr.Artifacts))
	d, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, "baidu.com", gjson.GetBytes(d, "url_parsed.domain").String())
	assert.Equal(t, "www", gjson.GetBytes(d, "url_parsed.subdomain").String())
	assert.Equal(t, "80", gjson.GetBytes(d, "url_parsed.port").String())
	assert.Equal(t, "/a/b/c.html", gjson.GetBytes(d, "url_parsed.path").String())
	assert.Equal(t, "/a/b", gjson.GetBytes(d, "url_parsed.dir").String())
	assert.Equal(t, "c.html", gjson.GetBytes(d, "url_parsed.file").String())
	assert.Equal(t, ".html", gjson.GetBytes(d, "url_parsed.ext").String())
	assert.False(t, gjson.GetBytes(d, "url_parsed.ip").Exists())
	fr = ParseURL(newTestRunner(t, data), map[string]interface{}{
		"parseIP": true,
	})
	assert.NotEqual(t, "", fr.OutFile)
	assert.Equal(t, 0, len(fr.Artifacts))
	d, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.True(t, gjson.GetBytes(d, "url_parsed.ip").Exists())

	data = `{"url":"httP://www.baidu.com"}`
	fr = ParseURL(newTestRunner(t, data), map[string]interface{}{})
	assert.NotEqual(t, "", fr.OutFile)
	assert.Equal(t, 0, len(fr.Artifacts))
	d, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, "baidu.com", gjson.GetBytes(d, "url_parsed.domain").String())
	assert.Equal(t, "www", gjson.GetBytes(d, "url_parsed.subdomain").String())
	assert.Equal(t, "80", gjson.GetBytes(d, "url_parsed.port").String())

	data = `{"url":"httP://baidu.com:81"}`
	fr = ParseURL(newTestRunner(t, data), map[string]interface{}{})
	assert.NotEqual(t, "", fr.OutFile)
	assert.Equal(t, 0, len(fr.Artifacts))
	d, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, "baidu.com", gjson.GetBytes(d, "url_parsed.domain").String())
	assert.Equal(t, "", gjson.GetBytes(d, "url_parsed.subdomain").String())
	assert.Equal(t, "81", gjson.GetBytes(d, "url_parsed.port").String())

	data = `{"url":"httP://1.1.1.1:81"}`
	fr = ParseURL(newTestRunner(t, data), map[string]interface{}{})
	assert.NotEqual(t, "", fr.OutFile)
	assert.Equal(t, 0, len(fr.Artifacts))
	d, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, "", gjson.GetBytes(d, "url_parsed.domain").String())
	assert.Equal(t, "1.1.1.1", gjson.GetBytes(d, "url_parsed.ip").String())
	assert.Equal(t, "", gjson.GetBytes(d, "url_parsed.subdomain").String())
	assert.Equal(t, "81", gjson.GetBytes(d, "url_parsed.port").String())
	assert.True(t, gjson.GetBytes(d, "url_parsed.ip").Exists())

	data = `{"url":"httPs://www.baidu.com"}`
	fr = ParseURL(newTestRunner(t, data), map[string]interface{}{})
	assert.NotEqual(t, "", fr.OutFile)
	assert.Equal(t, 0, len(fr.Artifacts))
	d, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, "baidu.com", gjson.GetBytes(d, "url_parsed.domain").String())
	assert.Equal(t, "www", gjson.GetBytes(d, "url_parsed.subdomain").String())
	assert.Equal(t, "443", gjson.GetBytes(d, "url_parsed.port").String())
	assert.Equal(t, "https", gjson.GetBytes(d, "url_parsed.scheme").String())

	//t.Log(string(d))

	data = `{"url":"http://www.cas.cn."}`
	fr = ParseURL(newTestRunner(t, data), map[string]interface{}{})
	assert.NotEqual(t, "", fr.OutFile)
	assert.Equal(t, 0, len(fr.Artifacts))
	d, err = utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, "", gjson.GetBytes(d, "url_parsed.domain").String())
	assert.Equal(t, "", gjson.GetBytes(d, "url_parsed.subdomain").String())
	assert.Equal(t, "", gjson.GetBytes(d, "url_parsed.port").String())
	assert.Equal(t, "", gjson.GetBytes(d, "url_parsed.scheme").String())
}
