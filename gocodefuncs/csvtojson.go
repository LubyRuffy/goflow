package gocodefuncs

import (
	"encoding/csv"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/sjson"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CSVToJsonParams 获取fofa的参数
type CSVToJsonParams struct {
}

func readCsv(f io.Reader) ([][]string, error) {
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readCsv(f)
}

// sjsonFileName 转换为 sjson可以处理的文件名
func sjsonFileName(fn string) string {
	filename := filepath.Base(fn)
	filename = strings.ReplaceAll(filename, ".", "\\.") // 坑：path会自动处理.符号，需要进行转义，否则扩展名就变成了子obj
	return filename
}

// CSVToJson 从csv读取内容到json
func CSVToJson(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options CSVToJsonParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("fetchFofa failed: %w", err))
	}

	if p.GetLastFile() == "" {
		panic("no file to read")
	}

	records, err := readCsvFile(p.GetLastFile())
	if err != nil {
		panic(fmt.Errorf("read csv failed: %w", err))
	}

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		line := ""
		line, err = sjson.Set(line, sjsonFileName(p.GetLastFile()), records)
		if err != nil {
			return err
		}

		_, err = f.WriteString(line)
		return err
	})
	if err != nil {
		panic(fmt.Errorf("read csv error: %w", err))
	}

	return &FuncResult{
		OutFile: fn,
	}
}
