package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPipeRunner_bar(t *testing.T) {
	assert.Equal(t,
		`BarChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "count()",
    "size": -1,
    "title": "",
})
`,
		workflowast.NewParser().MustParse(`bar("country")`))

	assert.Equal(t,
		`BarChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "size",
    "size": -1,
    "title": "",
})
`,
		workflowast.NewParser().MustParse(`bar("country", "size")`))

	assert.Equal(t,
		`BarChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "count()",
    "size": -1,
    "title": "",
})
`,
		workflowast.NewParser().MustParse(`bar("country", "")`))

	assert.Equal(t,
		`BarChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "size",
    "size": 5,
    "title": "",
})
`,
		workflowast.NewParser().MustParse(`bar("country", "size", 5)`))

	assert.Equal(t,
		`BarChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "size",
    "size": 2,
    "title": "test title",
})
`,
		workflowast.NewParser().MustParse(`bar("country", "size", 2, "test title")`))
}
