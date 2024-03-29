package gocodefuncs

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/saintfish/chardet"
	"github.com/tidwall/sjson"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// CSVToJsonParams 获取fofa的参数
type CSVToJsonParams struct {
}

func readCsv(f io.Reader) ([][]string, error) {
	var content bytes.Buffer
	var utf8Content []byte
	_, err := io.Copy(&content, f)
	if err != nil {
		return nil, err
	}
	if isGBK(content.Bytes()) {
		utf8Content, err = convertGBK2Str(content.Bytes())
		if err != nil {
			return nil, err
		}
	} else {
		utf8Content = content.Bytes()
	}

	csvReader := csv.NewReader(bytes.NewReader(utf8Content))
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func convertGBK2Str(gbkStr []byte) ([]byte, error) {
	//如果是[]byte格式的字符串，可以使用Bytes方法
	b, err := simplifiedchinese.GBK.NewDecoder().Bytes([]byte(gbkStr))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func isGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		//fmt.Printf("for %x\n", data[i])
		if data[i] <= 0xff {
			//编码小于等于127,只有一个字节的编码，兼容ASCII吗
			i++
			continue
		} else {
			//大于127的使用双字节编码
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readCsv(f)
}

func trFn(s string, cs string) string {
	switch strings.ToLower(cs) {
	case "gb-18030", "gbk":
		if utf8Data, err := simplifiedchinese.GBK.NewDecoder().String(s); err == nil {
			return string(utf8Data)
		}
	case "big5":
		if utf8Data, err := traditionalchinese.Big5.NewDecoder().String(s); err == nil {
			return string(utf8Data)
		}
	case "unicode.bigendian":
		if utf8Data, err := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder().String(s); err == nil {
			return string(utf8Data)
		}
	}
	return s
}

func tryToUtf8(s string, charset string) string {
	if !utf8.Valid([]byte(s)) {
		if len(charset) > 0 {
			return trFn(s, charset)
		}

		d := chardet.NewTextDetector()
		all, err := d.DetectAll([]byte(s))
		if err != nil {
			return string(s)
		}

		return trFn(string(s), all[0].Charset)

		//if isGBK(s) {
		//	utf8Data, _ := simplifiedchinese.GBK.NewDecoder().Bytes(s)
		//	return string(utf8Data)
		//}
		//
		//if v, _, err := transform.Bytes(unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder(), s); err == nil {
		//	if utf8.Valid(v) {
		//		return string(v)
		//	}
		//} else if v, _, err := transform.Bytes(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder(), s); err == nil {
		//	if utf8.Valid(v) {
		//		return string(v)
		//	}
		//}
	}
	return string(s)
}

// sjsonFileName 转换为 sjson可以处理的文件名
func sjsonFileName(fn string) string {
	fn = tryToUtf8(fn, "")
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
		//line, err = sjson.Set(line, sjsonFileName(p.GetLastFile()), records)
		line, err = sjson.Set(line, "Sheet1", records)
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
