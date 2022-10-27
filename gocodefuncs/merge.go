package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/mitchellh/mapstructure"
	"os"
)

type mergeParams struct {
	File string
}

// Merge 合并文件内容，相当于union
func Merge(p Runner, params map[string]interface{}) *FuncResult {

	var err error
	var options mergeParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(err)
	}

	emptyfr := &FuncResult{
		OutFile:   "",
		Artifacts: nil,
	}

	// 没有文件
	if p.GetLastFile() == "" {
		return emptyfr
	}

	var f1Data, f2Data []byte
	if options.File != "" {
		f1Data, err = os.ReadFile(options.File)
		if err != nil {
			panic(err)
		}
	}

	if p.GetLastFile() != "" {
		f2Data, err = os.ReadFile(p.GetLastFile())
		if err != nil {
			panic(err)
		}
	}

	fn, err := utils.WriteTempFile(".json", func(f *os.File) error {
		if f1Data != nil && len(f1Data) > 0 {
			_, err = f.Write(f1Data)
			if err != nil {
				return err
			}
			if f1Data[len(f1Data)-1] != '\n' {
				_, err = f.WriteString("\n")
				if err != nil {
					return err
				}
			}
		}

		if f2Data != nil {
			_, err = f.Write(f2Data)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return &FuncResult{
		OutFile:   fn,
		Artifacts: nil,
	}
}
