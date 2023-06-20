package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFileLines(t *testing.T) {
	fn, err := WriteTempFile(".txt", func(f *os.File) error {
		_, err := f.WriteString("aaa")
		return err
	})
	assert.Nil(t, err)
	assert.FileExists(t, fn)
	n, err := FileLines(fn)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), n)

	fn, err = WriteTempFile(".txt", func(f *os.File) error {
		_, err := f.WriteString("aaa\nbbb")
		return err
	})
	assert.Nil(t, err)
	assert.FileExists(t, fn)
	n, err = FileLines(fn)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), n)

	fn, err = WriteTempFile(".txt", func(f *os.File) error {
		v := ""
		for i := 0; i < 100000; i++ {
			v += fmt.Sprintf("-line %d\n", i)
		}
		_, err := f.WriteString(v)
		return err
	})
	assert.Nil(t, err)
	assert.FileExists(t, fn)
	n, err = FileLines(fn)
	assert.Nil(t, err)
	assert.Equal(t, int64(100000), n)
}

func TestTarGzFiles(t *testing.T) {
	fn1, err := WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"a":1}`)
		return err
	})
	assert.Nil(t, err)

	fn2, err := WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`{"b":1}`)
		return err
	})
	assert.Nil(t, err)

	tarGzData, err := TarGzFiles([]string{fn1, fn2})
	assert.Nil(t, err)

	// ungzip
	zr, err := gzip.NewReader(bytes.NewReader(tarGzData))
	assert.Nil(t, err)
	// untar
	tr := tar.NewReader(zr)

	i := 0
	// uncompress each element
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		assert.Nil(t, err)
		switch i {
		case 0:
			assert.Equal(t, filepath.Base(fn1), header.Name)
		case 1:
			assert.Equal(t, filepath.Base(fn2), header.Name)
		}
		i++
	}

}

func TestEachLineWithContext(t *testing.T) {
	i := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fn, err := WriteTempFile(".json", func(f *os.File) error {
		_, err := f.WriteString(`
{"a":1}
{"a":1}
{"a":1}
{"a":1}
{"a":1}
{"a":1}
`)
		return err
	})
	assert.Nil(t, err)
	assert.FileExists(t, fn)

	startCh := make(chan bool, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		EachLineWithContext(ctx, fn, func(line string) error {
			startCh <- true
			t.Log(line)
			time.Sleep(time.Second)
			i++
			return nil
		})
	}()
	<-startCh
	cancel()
	wg.Wait()

	assert.Equal(t, 1, i)
}

func Test_writeTempFile(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试写入文件",
			args: args{
				filename: "test.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := []byte("write temporary file sample")

			// 不根据文件名写入文件
			fn, err := WriteTempFile(filepath.Ext(tt.args.filename), func(f *os.File) error {
				_, err := f.Write(buf)
				return err
			})
			assert.Nil(t, err)
			assert.FileExists(t, fn)
			fn = filepath.Base(fn)
			match, err := regexp.Match(fmt.Sprintf(`%s\d+.json`, defaultPipeTmpFilePrefix), []byte(fn))
			assert.Nil(t, err)
			assert.Truef(t, match, fmt.Sprintf("unmatched filename: %s", fn))

			// 根据文件名写入文件
			fn, err = WriteTempFileWithName(tt.args.filename, func(f *os.File) error {
				_, err = f.Write(buf)
				return err
			})
			assert.Nil(t, err)
			assert.FileExists(t, fn)
			fn = filepath.Base(fn)
			match, err = regexp.Match(`\d+_test.json`, []byte(fn))
			assert.Nil(t, err)
			assert.Truef(t, match, fmt.Sprintf("unmatched filename: %s", fn))

			// 只根据文件名写入文件，不添加随机前缀
			fn, err = WriteTempFileWithNameOnly(tt.args.filename, func(f *os.File) error {
				_, err = f.Write(buf)
				return err
			})
			assert.Nil(t, err)
			assert.FileExists(t, fn)
			fn = filepath.Base(fn)
			match = strings.EqualFold(fn, tt.args.filename)
			assert.Nil(t, err)
			assert.Truef(t, match, fmt.Sprintf("unmatched filename: %s", fn))
		})
	}
}

func TestMoveFileTo(t *testing.T) {
	type args struct {
		src string
		dst string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "测试移动文件",
			args: args{
				src: "file_test_sample.bin",
				dst: "./test/test.bin",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, MoveFileTo(tt.args.src, tt.args.dst), fmt.Sprintf("MoveFileTo(%v, %v)", tt.args.src, tt.args.dst))
			assert.FileExists(t, tt.args.dst)
			tt.wantErr(t, MoveFileTo(tt.args.dst, tt.args.src), fmt.Sprintf("MoveFileTo(%v, %v)", tt.args.src, tt.args.dst))
		})
	}
}

func TestZipFiles(t *testing.T) {
	type args struct {
		files []string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "测试 zip 压缩文件",
			args:    args{files: []string{"./file_test_sample.bin"}},
			want:    []byte{0x50, 0x4b, 0x3, 0x4, 0x14, 0x0, 0x8, 0x0, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x16, 0x0, 0x0, 0x0, 0x2e, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x62, 0x69, 0x6e, 0x2a, 0x49, 0x2d, 0x2e, 0xe1, 0x2, 0x11, 0x80, 0x0, 0x0, 0x0, 0xff, 0xff, 0x50, 0x4b, 0x7, 0x8, 0x78, 0x14, 0xf5, 0x2d, 0xd, 0x0, 0x0, 0x0, 0x9, 0x0, 0x0, 0x0, 0x50, 0x4b, 0x1, 0x2, 0x14, 0x0, 0x14, 0x0, 0x8, 0x0, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x78, 0x14, 0xf5, 0x2d, 0xd, 0x0, 0x0, 0x0, 0x9, 0x0, 0x0, 0x0, 0x16, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2e, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x62, 0x69, 0x6e, 0x50, 0x4b, 0x5, 0x6, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x1, 0x0, 0x44, 0x0, 0x0, 0x0, 0x51, 0x0, 0x0, 0x0, 0x0, 0x0},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ZipFiles(tt.args.files)
			if !tt.wantErr(t, err, fmt.Sprintf("ZipFiles(%v)", tt.args.files)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ZipFiles(%v)", tt.args.files)
			fn, _ := WriteTempFile(".zip", func(f *os.File) error {
				f.Write(got)
				return nil
			})
			log.Printf("zip file write to: %s", fn)
		})
	}
}
