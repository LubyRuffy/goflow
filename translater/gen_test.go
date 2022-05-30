package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad_gen(t *testing.T) {
	assert.Equal(t,
		`GenData(GetRunner(), map[string]interface{} {
    "data": "{\"a\":\"json\"}",
})
`,
		workflowast.NewParser().MustParse(`gen("{\"a\":\"json\"}")`))
}
