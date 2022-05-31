package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPipeRunner_urlfix(t *testing.T) {
	assert.Equal(t,
		"URLFix(GetRunner(), map[string]interface{}{\n    \"urlField\": \"url\",\n})\n",
		workflowast.NewParser().MustParse(`fix_url()`))
	assert.Equal(t,
		"URLFix(GetRunner(), map[string]interface{}{\n    \"urlField\": \"host\",\n})\n",
		workflowast.NewParser().MustParse(`fix_url("host")`))

}
