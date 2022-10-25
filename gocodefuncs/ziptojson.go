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

func doZipToJson(f string) string {
	zf, err := zip.OpenReader(f)
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

			filename := tryToUtf8(file.Name, "gbk") // utf8
			records, err := loadFileToJson(filename, fc)
			if err != nil {
				//p.Warnf("read zip file failed: ", err)
				continue
			}

			line, err = sjson.Set(line, sjsonFileName(filename), records)
			if err != nil {
				return err
			}
		}

		_, err = f.WriteString(line)
		return err
	})
	if err != nil {
		panic(fmt.Errorf("read zip error: %w", err))
	}
	return fn
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

	return &FuncResult{
		OutFile: doZipToJson(p.GetLastFile()),
	}
}
