package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad_httpGet(t *testing.T) {
	assert.Equal(t,
		`HttpRequest(GetRunner(), map[string]interface{} {
    "urlField": "url",
    "userAgent": "",
    "tlsVerify": false,
    "workers": 5,
    "maxSize": -1,
})
`,
		workflowast.NewParser().MustParse(`http_get()`))

	assert.Equal(t,
		`HttpRequest(GetRunner(), map[string]interface{} {
    "urlField": "host_url",
    "userAgent": "",
    "tlsVerify": false,
    "workers": 5,
    "maxSize": -1,
})
`,
		workflowast.NewParser().MustParse(`http_get("host_url")`))

	assert.Equal(t,
		`HttpRequest(GetRunner(), map[string]interface{} {
    "urlField": "host_url",
    "userAgent": "my ua",
    "tlsVerify": false,
    "workers": 5,
    "maxSize": -1,
})
`,
		workflowast.NewParser().MustParse(`http_get("host_url", "my ua")`))

	assert.Equal(t,
		`HttpRequest(GetRunner(), map[string]interface{} {
    "urlField": "host_url",
    "userAgent": "my ua",
    "tlsVerify": true,
    "workers": 5,
    "maxSize": -1,
})
`,
		workflowast.NewParser().MustParse(`http_get("host_url", "my ua", true)`))

	assert.Equal(t,
		`HttpRequest(GetRunner(), map[string]interface{} {
    "urlField": "host_url",
    "userAgent": "my ua",
    "tlsVerify": true,
    "workers": 10,
    "maxSize": -1,
})
`,
		workflowast.NewParser().MustParse(`http_get("host_url", "my ua", true, 10)`))

	assert.Equal(t,
		`HttpRequest(GetRunner(), map[string]interface{} {
    "urlField": "host_url",
    "userAgent": "my ua",
    "tlsVerify": true,
    "workers": 10,
    "maxSize": 1000,
})
`,
		workflowast.NewParser().MustParse(`http_get("host_url", "my ua", true, 10, 1000)`))

}
