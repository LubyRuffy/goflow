package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad_load(t *testing.T) {
	assert.Equal(t,
		`LoadFile(GetRunner(), map[string]interface{} {
    "file": "test.json",
})
`,
		workflowast.NewParser().MustParse(`load("test.json")`))
}
