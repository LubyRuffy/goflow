package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"strconv"
)

/*
$  echo '{"ip":"1.1.1.1","port":"80"}' | ./zq.exe ' switch ( case has(ip) => put url:=ip+":"+port default => this) ' -
{ip:"1.1.1.1",port:"80",url:"1.1.1.1:80"}

$  echo '{"ip":"1.1.1.1","port":"80"}' | ./zq.exe ' switch ( case has(a) => put url:=ip+":"+port default => yield this)' -
{ip:"1.1.1.1",port:"80"}

if_add(`has("a")`, `url`, `ip+":"+port`)
*/
func ifAddHook(fi *workflowast.FuncInfo) string {
	q := strconv.Quote(`switch ( case ` + fi.Params[0].RawString() + ` => put ` + fi.Params[1].RawString() + `:=` + fi.Params[2].RawString() + ` default => yield this)`)
	return `ZqQuery(GetRunner(), map[string]interface{}{
    "query": ` + q + `,
})`
}

func init() {
	Register("if_add", ifAddHook) // 剪出要的字段
}
