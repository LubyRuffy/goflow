package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad_screenshot(t *testing.T) {
	assert.Equal(t,
		`Screenshot(GetRunner(), map[string]interface{}{
	"urlField": "host",
	"saveField": "screenshot.filepath",
	"timeout": 30,
	"workers": 5,
})
`,
		workflowast.NewParser().MustParse(`screenshot("host")`))

	assert.Equal(t,
		`Screenshot(GetRunner(), map[string]interface{}{
	"urlField": "url",
	"saveField": "screenshot.filepath",
	"timeout": 30,
	"workers": 5,
})
`,
		workflowast.NewParser().MustParse(`screenshot()`))

	assert.Equal(t,
		`Screenshot(GetRunner(), map[string]interface{}{
	"urlField": "url",
	"saveField": "screenshot.filepath",
	"timeout": 30,
	"workers": 5,
})
`,
		workflowast.NewParser().MustParse(`screenshot("")`))

	assert.Equal(t,
		`Screenshot(GetRunner(), map[string]interface{}{
	"urlField": "host",
	"saveField": "sc_filepath",
	"timeout": 30,
	"workers": 5,
})
`,
		workflowast.NewParser().MustParse(`screenshot("host", "sc_filepath")`))

	assert.Equal(t,
		`Screenshot(GetRunner(), map[string]interface{}{
	"urlField": "host",
	"saveField": "sc_filepath",
	"timeout": 1,
	"workers": 5,
})
`,
		workflowast.NewParser().MustParse(`screenshot("host", "sc_filepath", 1)`))

	assert.Equal(t,
		`Screenshot(GetRunner(), map[string]interface{}{
	"urlField": "host",
	"saveField": "sc_filepath",
	"timeout": 1,
	"workers": 10,
})
`,
		workflowast.NewParser().MustParse(`screenshot("host", "sc_filepath", 1, 10)`))

}
