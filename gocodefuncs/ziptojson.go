package gocodefuncs

import (
	"archive/zip"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/sjson"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipToJsonParams 获取fofa的参数
type ZipToJsonParams struct {
}

func loadFileToJson(name string, f io.Reader) (interface{}, error) {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".csv":
		return readCsv(f)
	case ".xlsx":
		return readExcel(f)
	}
	return nil, nil
}

// ZipToJson 从csv读取内容到json
func ZipToJson(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options CSVToJsonParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("fetchFofa failed: %w", err))
	}

	if p.GetLastFile() == "" {
		panic("no file to read")
	}

	zf, err := zip.OpenReader(p.GetLastFile())
	if err != nil {
		panic(fmt.Errorf("read zip file failed: %w", err))
	}
	defer zf.Close()

	var fn string
	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		line := ""

		for _, file := range zf.File {
			fc, err := file.Open()
			defer fc.Close()

			records, err := loadFileToJson(file.Name, fc)
			if err != nil {
				p.Warnf("read zip file failed: ", err)
				continue
				//panic(fmt.Errorf("read zip file failed: %w", err))
			}

			line, err = sjson.Set(line, sjsonFileName(file.Name), records)
			if err != nil {
				return err
			}
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
