package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPipeRunner_where(t *testing.T) {
	// 是否存在字段
	assert.Equal(t,
		"ZqQuery(GetRunner(), map[string]interface{}{\n    \"query\": \"where has(a)\",\n})\n",
		workflowast.NewParser().MustParse(`where("has(a)")`))

	// 字段值的长度大于
	assert.Equal(t,
		"ZqQuery(GetRunner(), map[string]interface{}{\n    \"query\": \"where len(a)>100\",\n})\n",
		workflowast.NewParser().MustParse(`where("len(a)>100")`))
}
