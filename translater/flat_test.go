package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPipeRunner_flat(t *testing.T) {
	assert.Equal(t,
		`FlatArray(GetRunner(), map[string]interface{}{
    "field": "a",
})
`,
		workflowast.NewParser().MustParse(`flat("a")`))
}
