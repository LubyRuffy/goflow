package utils

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	defaultPipeTmpFilePrefix = "goflow_pipeline_"
)

// EachLineWithContext 支持中止的方式
func EachLineWithContext(ctx context.Context, filename string, f func(line string) error) error {
	return EachLine(filename, func(line string) error {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		return f(line)
	})
}

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

// writeTempFile 向Temp文件夹写入文件
// 如果writeF是nil，就只返回生成的一个临时空文件路径
// 返回文件名和错误
func writeTempFile(filename string, writeF func(f *os.File) error) (fn string, err error) {
	var f *os.File
	if len(filename) > 0 && filepath.Ext(filename) == filename || filename == "" {
		filename = "*" + filename
		f, err = os.CreateTemp(os.TempDir(), defaultPipeTmpFilePrefix+filename)
	} else {
		if strings.Contains(filename, "*") {
			f, err = os.CreateTemp(os.TempDir(), filename)
		} else {
			f, err = os.Create(filepath.Join(os.TempDir(), filename))
		}
	}

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

// WriteTempFile 写入临时文件
// 输入需要写入的后缀
// 如果writeF是nil，就只返回生成的一个临时空文件路径
// 返回文件名和错误
func WriteTempFile(ext string, writeF func(f *os.File) error) (fn string, err error) {
	return writeTempFile(ext, writeF)
}

// WriteTempFileWithNameOnly 写入临时文件
// 输入文件名称
// 如果writeF是nil，就只返回生成的一个临时空文件路径
// 返回文件名和错误
func WriteTempFileWithNameOnly(filename string, writeF func(f *os.File) error) (fn string, err error) {
	return writeTempFile(filename, writeF)
}

// WriteTempFileWithName 写入临时文件，指定生成的文件名
// 输入需要写入的文件名
// 如果writeF是nil，就只返回生成的一个临时空文件路径
// 返回文件名和错误
func WriteTempFileWithName(filename string, writeF func(f *os.File) error) (fn string, err error) {
	return writeTempFile("*_"+filename, writeF)
}

// MoveFileTo 移动文件至指定位置
func MoveFileTo(src string, dst string) (err error) {
	path := filepath.Dir(dst)
	exist := false
	// 判断文件夹是否存在
	_, err = os.Stat(path)
	if err == nil {
		exist = true
	}
	if os.IsNotExist(err) {
		exist = false
	}

	// 创建文件夹
	if !exist {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return os.Rename(src, dst)
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

// TarGzFiles 打包文件为tar.gz
func TarGzFiles(files []string) ([]byte, error) {
	var buf bytes.Buffer
	zr := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zr)
	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			log.Println("load file failed:", file, err)
			continue
		}
		header, err := tar.FileInfoHeader(fi, file)
		header.Name = filepath.Base(file)
		// write header
		if err = tw.WriteHeader(header); err != nil {
			log.Println("compress file failed:", file, err)
			continue
		}

		data, err := os.Open(file)
		if err != nil {
			log.Println("compress file failed:", file, err)
			continue
		}
		if _, err = io.Copy(tw, data); err != nil {
			log.Println("compress file failed:", file, err)
			continue
		}
	}

	// produce tar
	if err := tw.Close(); err != nil {
		return nil, err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ZipFiles(files []string) ([]byte, error) {
	// 创建输出文件
	outFile, err := os.CreateTemp(os.TempDir(), defaultPipeTmpFilePrefix+"*.zip")
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)

	for _, file := range files {
		// 这里压缩文件的路径会默认取 ./filename ,需要改为 base filename： filepath.Base(file)
		zipEntry, err := zipWriter.Create(filepath.Base(file))
		if err != nil {
			log.Printf("cannot create zipEntry when zip file: %s, error: %s", file, err.Error())
			continue
		}

		inFile, err := os.Open(file)
		if err != nil {
			log.Printf("zip file: fail to open target file %s, error: %s", file, err.Error())
			continue
		}
		defer inFile.Close()

		_, err = io.Copy(zipEntry, inFile)
		if err != nil {
			log.Printf("zip file: fail to copy entry for target file %s, error: %s", file, err.Error())
			continue
		}
	}

	// produce tar
	if err = zipWriter.Close(); err != nil {
		return nil, err
	}

	err = outFile.Close()
	if err != nil {
		return nil, err
	}

	fileContent, err := os.ReadFile(outFile.Name())
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}
