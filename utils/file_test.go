package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFileLines(t *testing.T) {
	fn, err := WriteTempFile(".txt", func(f *os.File) error {
		v := ""
		for i := 0; i < 100000; i++ {
			v += fmt.Sprintf("-line %d\n", i)
		}
		_, err := f.WriteString(v)
		return err
	})
	assert.Nil(t, err)
	assert.FileExists(t, fn)
	n, err := FileLines(fn)
	assert.Nil(t, err)
	assert.Equal(t, int64(100000), n)
}
