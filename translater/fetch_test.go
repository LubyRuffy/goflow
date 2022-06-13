package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad_fetch(t *testing.T) {
	assert.Equal(t,
		`FetchFile(GetRunner(), map[string]interface{} {
    "url": "test.json",
})
`,
		workflowast.NewParser().MustParse(`fetch("test.json")`))
}
