package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPipeRunner_pie(t *testing.T) {
	assert.Equal(t,
		`PieChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "count()",
    "size": -1,
    "title": "",
})
`,
		workflowast.NewParser().MustParse(`pie("country")`))

	assert.Equal(t,
		`PieChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "size",
    "size": -1,
    "title": "",
})
`,
		workflowast.NewParser().MustParse(`pie("country", "size")`))

	assert.Equal(t,
		`PieChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "count()",
    "size": -1,
    "title": "",
})
`,
		workflowast.NewParser().MustParse(`pie("country", "")`))

	assert.Equal(t,
		`PieChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "size",
    "size": 5,
    "title": "",
})
`,
		workflowast.NewParser().MustParse(`pie("country", "size", 5)`))

	assert.Equal(t,
		`PieChart(GetRunner(), map[string]interface{}{
    "name": "country",
    "value": "size",
    "size": 2,
    "title": "test title",
})
`,
		workflowast.NewParser().MustParse(`pie("country", "size", 2, "test title")`))
}
