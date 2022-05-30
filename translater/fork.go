package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
)

func forkHook(fi *workflowast.FuncInfo) string {
	return `Fork(` + fi.Params[0].String() + `)`
}

func init() {
	register("fork", forkHook) // fork
}
