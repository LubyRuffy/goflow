package translater

import (
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad_to_sqlite(t *testing.T) {
	assert.Panics(t, func() {
		workflowast.NewParser().MustParse(`to_sqlite()`)
	})

	assert.Equal(t,
		`ToSql(GetRunner(), map[string]interface{}{
	"driver": "sqlite3",
	"table": "tbl",
	"fields": "",
})
`,
		workflowast.NewParser().MustParse(`to_sqlite("tbl")`))

	assert.Equal(t,
		`ToSql(GetRunner(), map[string]interface{}{
	"driver": "sqlite3",
	"table": "tbl1",
	"fields": "a,b,c",
})
`,
		workflowast.NewParser().MustParse(`to_sqlite("tbl1", "a,b,c")`))
}
