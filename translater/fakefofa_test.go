package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad_fakefofa(t *testing.T) {
	assert.Equal(t,
		`GenFofaFieldData(GetRunner(), map[string]interface{} {
    "query": "host=\"https://fofa.info\"",
    "size": 1,
    "fields": "domain",
})
`,
		workflowast.NewParser().MustParse(`fakefofa("host=\"https://fofa.info\"", "domain", 1)`))
}
