package gocodefuncs

import (
	"github.com/LubyRuffy/goflow/utils"
	"github.com/itchyny/gojq/cli"
	"github.com/mitchellh/mapstructure"
	"log"
	"os"
)

type jqQueryParams struct {
	Query  string `json:"query"`
	Stream bool
}

// JqQuery jq command
func JqQuery(p Runner, params map[string]interface{}) *FuncResult {
	var fn string
	var err error
	var options jqQueryParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(err)
	}

	inFile, err := os.Open(p.GetLastFile())
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	args := []string{os.Args[0]}
	if options.Stream {
		args = append(args, "-s")
	}
	origArgs := os.Args
	defer func() {
		os.Args = origArgs
	}()
	os.Args = append(args, "-c", options.Query)

	outR, outW, _ := os.Pipe()
	origStdout := os.Stdout
	defer func() {
		os.Stdout = origStdout
	}()
	os.Stdout = outW

	origStdin := os.Stdin
	defer func() {
		os.Stdin = origStdin
	}()
	os.Stdin = inFile

	r, w, _ := os.Pipe()
	os.Stderr = w

	cli.Run()

	ch := make(chan string, 1)
	go func() {
		defer func() {
			ch <- fn
		}()
		fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
			for {
				var buf [1024]byte
				n, err := outR.Read(buf[:])
				if err != nil {
					break
				}
				if n <= 0 {
					break
				}
				_, err = f.Write(buf[:n])
				if err != nil {
					return err
				}
				if n < 1024 {
					break
				}
			}
			return nil
		})
	}()

	go func() {
		defer func() {
			ch <- ""
		}()
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(buf[:n]))
	}()

	select {
	case <-ch:
	}

	close(ch)

	return &FuncResult{
		OutFile: fn,
	}
}
