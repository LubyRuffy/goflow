package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/brimdata/zed/cli/zq"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// downloadFile 下载文件
func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// FetchFileParams 获取fofa的参数
type FetchFileParams struct {
	URL    string // url
	Format string // 格式，如果有，直接调用解析器，如果没有，根据文件后缀进行
}

// FetchFile 从网络文件获取数据
func FetchFile(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options FetchFileParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(fmt.Errorf("FetchFile failed: %w", err))
	}

	if len(options.URL) == 0 {
		panic(fmt.Errorf("url cannot be empty"))
	}
	u, err := url.Parse(options.URL)
	if err != nil {
		panic(fmt.Errorf("FetchFile failed: %w", err))
	}

	// 存储原始文件
	var rawFile string
	rawFile, err = utils.WriteTempFile(filepath.Ext(u.Path), nil)
	if err != nil {
		panic(fmt.Errorf("FetchFile failed: %w", err))
	}
	if err = downloadFile(rawFile, options.URL); err != nil {
		panic(err)
	}

	// 转文件为json，必须是结构话的
	if options.Format == "" {
		options.Format = filepath.Ext(u.Path)
	}

	var jsonFn string

	switch options.Format {
	case "csv", ".csv":
		jsonFn, err = utils.WriteTempFile(".json", nil)
		if err != nil {
			panic(fmt.Errorf("FetchFile failed: %w", err))
		}

		cmd := []string{"-f", "json", "-o", jsonFn, rawFile}
		err = zq.Cmd.ExecRoot(cmd)
		if err != nil {
			panic(fmt.Errorf("zqQuery error: %w", err))
		}
	case "json", ".json":
		jsonFn = rawFile
		var d []byte
		d, err = utils.ReadFirstLineOfFile(rawFile)
		if err != nil {
			panic(fmt.Errorf("FetchFile failed: %w", err))
		}

		// 第一行格式不对，说明是大的json数组
		if !gjson.ValidBytes(d) {
			d, _ = os.ReadFile(rawFile)
			result := gjson.ParseBytes(d)

			jsonFn, err = utils.WriteTempFile(".json", func(f *os.File) error {
				for _, a := range result.Array() {
					_, err = f.WriteString(a.Get("@ugly").String() + "\n")
					if err != nil {
						return err
					}
				}
				return nil
			})

			if err != nil {
				panic(fmt.Errorf("FetchFile failed: %w", err))
			}
		}

	default:
		panic(fmt.Errorf("unknown format of:%s", options.Format))
	}

	return &FuncResult{
		OutFile: jsonFn,
		Artifacts: []*Artifact{
			{
				FileName: filepath.Base(rawFile),
				FilePath: rawFile,
				Memo:     "raw download file",
			},
		},
	}
}

/*
- json有两种格式，需要进行归一化：一种是每行一条；一种是数组形式
*/
