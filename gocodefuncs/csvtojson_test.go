package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"os"
	"testing"
)

func writeSampleCSVFile() string {
	filename, err := utils.WriteTempFile(".csv", func(f *os.File) error {
		_, err := f.WriteString("a,b\n1,2")
		return err
	})
	if err != nil {
		panic(err)
	}
	return filename
}

func TestCSVToJson(t *testing.T) {
	filename := writeSampleCSVFile()
	fr := CSVToJson(&testRunner{
		T:        t,
		lastFile: filename,
	}, map[string]interface{}{})
	f, err := utils.ReadFirstLineOfFile(fr.OutFile)
	assert.Nil(t, err)
	gjson.ParseBytes(f).ForEach(func(key, value gjson.Result) bool {
		assert.Equal(t, key.String(), "Sheet1")
		assert.Equal(t, value.String(), `[["a","b"],["1","2"]]`)
		return false
	})

}

func Test_readCsvFile(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    [][]string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "csv读取中文乱码修复",
			args:    args{filePath: "./sample1.csv"},
			want:    [][]string{{"country", "city"}, {"美国", "北京"}},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readCsvFile(tt.args.filePath)
			if !tt.wantErr(t, err, fmt.Sprintf("readCsvFile(%v)", tt.args.filePath)) {
				return
			}
			assert.Equalf(t, tt.want, got, "readCsvFile(%v)", tt.args.filePath)
		})
	}
}
