package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
$  echo '{"ip":"1.1.1.1","port":"80"}' | ./zq.exe ' switch ( case has(ip) => put url:=ip+":"+port default => this) ' -
{ip:"1.1.1.1",port:"80",url:"1.1.1.1:80"}

$  echo '{"ip":"1.1.1.1","port":"80"}' | ./zq.exe ' switch ( case has(a) => put url:=ip+":"+port default => yield this)' -
{ip:"1.1.1.1",port:"80"}
*/
func TestPipeRunner_if_add(t *testing.T) {
	assert.Equal(t,
		"ZqQuery(GetRunner(), map[string]interface{}{\n    \"query\": \"switch ( case has(a) => put c:=a+\\\":\\\"+b default => yield this)\",\n})\n",
		workflowast.NewParser().MustParse(`if_add("has(a)", "c", "a+\":\"+b")`))
}
