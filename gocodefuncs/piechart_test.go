package gocodefuncs

import (
	"context"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
)

type testRunner struct {
	lastFile string
	objects  sync.Map
	*testing.T
}

func (t *testRunner) GetObject(name string) (interface{}, bool) {
	return t.objects.Load(name)
}
func (t *testRunner) SetObject(name string, v interface{}) {
	t.objects.Store(name, v)
}
func (t *testRunner) GetLastFile() string {
	return t.lastFile
}
func (t *testRunner) Debugf(format string, args ...interface{}) {
}
func (t *testRunner) Warnf(format string, args ...interface{}) {
}
func (t *testRunner) Logf(level logrus.Level, format string, args ...interface{}) {
}
func (t *testRunner) SetProgress(v float64) {
	t.T.Logf("progress: %f%%", 100*v)
}
func (t *testRunner) GetContext() context.Context {
	return context.Background()
}
func (t *testRunner) FormatResourceFieldInJson(filename string) (fn string, err error) {
	return filename, nil
}
func (t *testRunner) OnJobStart() {
	return
}
func (t *testRunner) OnJobFinished() {
	return
}
func (t *testRunner) LastFileEmpty() bool {
	return false
}

func newTestRunner(t *testing.T, jsonData string) *testRunner {
	fn, err := utils.WriteTempFile(".json", func(f *os.File) error {
		f.WriteString(jsonData)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return &testRunner{
		T:        t,
		lastFile: fn,
	}
}

func TestPieChart(t *testing.T) {
	data := `{"name":"a"}
{"name":"b"}
{"name":"a"}
{"name":"b"}
{"name":"b"}`
	fr := PieChart(newTestRunner(t, data), map[string]interface{}{
		"name":  "name",
		"value": "count()",
	})
	assert.Equal(t, 1, len(fr.Artifacts))
	assert.FileExists(t, fr.Artifacts[0].FilePath)
	d, err := os.ReadFile(fr.Artifacts[0].FilePath)
	assert.Nil(t, err)
	assert.Contains(t, string(d), `{"name":"b","value":3},{"name":"a","value":2}`)

	// 有value的相加测试
	data = `{"name":"a","size":1}
{"name":"b","size":2}
{"name":"a","size":3}
{"name":"b","size":4}
{"name":"b","size":5}`
	fr = PieChart(newTestRunner(t, data), map[string]interface{}{
		"name":  "name",
		"value": "size",
	})
	assert.Equal(t, 1, len(fr.Artifacts))
	assert.FileExists(t, fr.Artifacts[0].FilePath)
	d, err = os.ReadFile(fr.Artifacts[0].FilePath)
	assert.Nil(t, err)
	assert.Contains(t, string(d), `{"name":"b","value":11},{"name":"a","value":4}`)
}
