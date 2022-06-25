package utils

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

var (
	defaultPipeTmpFilePrefix = "goflow_pipeline_"
)

// EachLine 每行处理文件
func EachLine(filename string, f func(line string) error) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := read(reader)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		err = f(string(line))
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadFirstLineOfFile 读取文件的第一行
func ReadFirstLineOfFile(fn string) ([]byte, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var b [1]byte
	var data []byte
	for {
		_, err = f.Read(b[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return data, err
		}
		if b[0] == '\n' {
			break
		}
		data = append(data, b[0])
	}
	return data, nil
}

// FileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// LoadFirstExistsFile 从文件列表中返回第一个存在的文件路径
func LoadFirstExistsFile(paths []string) string {
	for _, p := range paths {
		if FileExists(p) {
			return p
		}
	}
	return ""
}

// WriteTempFile 写入临时文件
// 如果writeF是nil，就只返回生成的一个临时空文件路径
// 返回文件名和错误
func WriteTempFile(ext string, writeF func(f *os.File) error) (fn string, err error) {
	var f *os.File
	if len(ext) > 0 {
		ext = "*" + ext
	}
	f, err = os.CreateTemp(os.TempDir(), defaultPipeTmpFilePrefix+ext)
	if err != nil {
		return
	}
	defer f.Close()

	fn = f.Name()

	if writeF != nil {
		err = writeF(f)
		if err != nil {
			return
		}
	}
	return
}

// FileLines 统计文件行
func FileLines(fileName string) (int64, error) {
	file, err := os.Open(fileName)

	if err != nil {
		return 0, err
	}

	buf := make([]byte, 1024)
	var lines int64
	var lastBytes int

	for {
		readBytes, err := file.Read(buf)
		if err != nil {
			if readBytes == 0 && err == io.EOF {
				err = nil
			}

			if l := len(buf[:lastBytes]); l > 0 && buf[l-1] != '\n' {
				lines++
			}
			return lines, err
		}

		lines += int64(bytes.Count(buf[:readBytes], []byte{'\n'}))
		lastBytes = readBytes
	}
}
