package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
$  echo '{"a":"1","b":"2"}' | ./zq.exe 'put c:=a+":"+b' -
{a:"1",b:"2",c:"1:2"}
*/
func TestPipeRunner_concat_add(t *testing.T) {
	assert.Equal(t,
		"ZqQuery(GetRunner(), map[string]interface{}{\n    \"query\": \"put c:=a+\\\":\\\"+b\",\n})\n",
		workflowast.NewParser().MustParse(`concat_add("a+\":\"+b", "c")`))
}
