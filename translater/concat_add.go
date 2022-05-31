package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"strconv"
)

func concatAddHook(fi *workflowast.FuncInfo) string {
	return `ZqQuery(GetRunner(), map[string]interface{}{
    "query": ` + strconv.Quote(`yield put `+fi.Params[1].RawString()+`:=`+fi.Params[0].RawString()) + `,
})`
}

func init() {
	Register("concat_add", concatAddHook) // grep匹配再新增字段
}
