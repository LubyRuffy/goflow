package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad_to_excel(t *testing.T) {
	assert.Equal(t,
		`ToExcel(GetRunner(), map[string]interface{}{
})
`,
		workflowast.NewParser().MustParse(`to_excel()`))

}
