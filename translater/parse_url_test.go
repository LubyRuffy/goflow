package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad_parse_url(t *testing.T) {
	assert.Equal(t,
		`ZqQuery(GetRunner(), map[string]interface{}{
	"query": "yield {...this, url_parsed:parse_uri(url)}",
})
`,
		workflowast.NewParser().MustParse(`parse_url()`))

	assert.Equal(t,
		`ZqQuery(GetRunner(), map[string]interface{}{
	"query": "yield {...this, host_parsed:parse_uri(host)}",
})
`,
		workflowast.NewParser().MustParse(`parse_url("host")`))
}
