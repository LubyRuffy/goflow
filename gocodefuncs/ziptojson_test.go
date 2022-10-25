package gocodefuncs

import (
	"archive/zip"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func writeSampleZipFile(t *testing.T) string {
	filename, err := utils.WriteTempFile(".zip", func(f *os.File) error {
		writer := zip.NewWriter(f)

		w1, _ := writer.Create("pom.csv")
		io.Copy(w1, strings.NewReader("a,b\n1,2"))

		w2, _ := writer.Create("testdir/Book1.xlsx")
		ef := excelize.NewFile()
		// Create a new worksheet.
		index := ef.NewSheet("Sheet2")
		// Set value of a cell.
		ef.SetCellValue("Sheet2", "A2", "Hello world.")
		ef.SetCellValue("Sheet1", "A1", "IP")
		ef.SetCellValue("Sheet1", "B1", "域名")
		ef.SetCellValue("Sheet1", "A2", "1.1.1.1")
		ef.SetCellValue("Sheet1", "B2", "a.com")
		// Set the active worksheet of the workbook.
		ef.SetActiveSheet(index)
		ef.Write(w2)
		ef.Close()

		// unicode
		bs_UTF16LE, _, _ := transform.Bytes(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder(), []byte("測試.csv"))
		w3, _ := writer.Create(string(bs_UTF16LE))
		io.Copy(w3, strings.NewReader("a,b\n1,2"))

		return writer.Close()
	})
	assert.Nil(t, err)
	t.Log(filename)
	return filename
}

func TestZipToJson(t *testing.T) {
	bs_UTF16LE, _, _ := transform.Bytes(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder(), []byte("測試"))
	assert.False(t, utf8.Valid(bs_UTF16LE))

	filename := writeSampleZipFile(t)
	fr := ZipToJson(&testRunner{
		T:        t,
		lastFile: filename,
	}, map[string]interface{}{})
	d, err := os.ReadFile(fr.OutFile)
	assert.Nil(t, err)
	assert.Equal(t, string(d), `{"pom.csv":[["a","b"],["1","2"]],"Book1.xlsx":{"Sheet1":[["IP","域名"],["1.1.1.1","a.com"]],"Sheet2":[null,["Hello world."]]},"測試.csv":[["a","b"],["1","2"]]}`)

}
