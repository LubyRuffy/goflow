package gocodefuncs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/itchyny/gojq"
	"github.com/mitchellh/mapstructure"
	"os"
)

type jqQueryParams struct {
	Query  string `json:"query"`
	Stream bool
}

func doJqQuery(inFile *os.File, options jqQueryParams, onData func([]byte)) error {
	q, err := gojq.Parse(options.Query)
	if err != nil {
		return err
	}
	code, err := gojq.Compile(q, gojq.WithInputIter(nil))
	if err != nil {
		return err
	}

	run := func(v interface{}, h func([]byte)) error {
		var iter gojq.Iter
		iter = code.Run(v)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok = v.(error); ok {
				return err
			}
			d, err := json.Marshal(v)
			if err != nil {
				return err
			}
			h(d)
		}
		return nil
	}

	var lines [][]byte
	decoder := json.NewDecoder(inFile)
	if options.Stream {
		var inputs []interface{}
		for decoder.More() {
			var oneJson interface{}
			if err := decoder.Decode(&oneJson); err != nil {
				return fmt.Errorf("parse order: %w", err)
			}
			inputs = append(inputs, oneJson)
		}
		if err = run(inputs, func(d []byte) {
			lines = append(lines, d)
		}); err != nil {
			return err
		}
	} else {
		for decoder.More() {
			var oneJson interface{}
			if err := decoder.Decode(&oneJson); err != nil {
				return fmt.Errorf("parse order: %w", err)
			}
			if err = run(oneJson, func(d []byte) {
				lines = append(lines, d)
			}); err != nil {
				return err
			}
		}
	}
	onData(bytes.Join(lines, []byte("\n")))

	return nil
}

//func doJqQuery(inFile *os.File, options jqQueryParams, onData func([]byte)) error {
//	args := []string{os.Args[0]}
//	if options.Stream {
//		args = append(args, "-s")
//	}
//	origArgs := os.Args
//	defer func() {
//		os.Args = origArgs
//	}()
//	os.Args = append(args, "-c", options.Query)
//
//	outR, outW, _ := os.Pipe()
//	defer func() {
//		outR.Close()
//		outW.Close()
//	}()
//	origStdout := os.Stdout
//	defer func() {
//		os.Stdout = origStdout
//	}()
//	os.Stdout = outW
//
//	origStdin := os.Stdin
//	defer func() {
//		os.Stdin = origStdin
//	}()
//	os.Stdin = inFile
//
//	errR, errW, _ := os.Pipe()
//	defer func() {
//		errR.Close()
//		errW.Close()
//	}()
//	os.Stderr = errW
//
//	errCh := make(chan error, 2)
//	var wg sync.WaitGroup
//	wg.Add(2)
//	go func() {
//		defer wg.Done()
//		buf, err := ioutil.ReadAll(outR)
//		if err != nil {
//			errCh <- err
//			return
//		}
//
//		onData(buf)
//	}()
//
//	go func() {
//		defer wg.Done()
//		buf, err := ioutil.ReadAll(errR)
//		if err != nil {
//			errCh <- err
//			return
//		}
//
//		if len(buf) > 0 {
//			errCh <- errors.New(string(buf))
//		}
//	}()
//
//	status := cli.Run()
//	errW.Close()
//	outW.Close()
//
//	wg.Wait()
//	close(errCh)
//
//	if status == 0 {
//		//return nil
//	}
//
//	for e := range errCh {
//		panic(e)
//	}
//
//	return nil
//}

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

	err = doJqQuery(inFile, options, func(buf []byte) {
		fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
			_, err = f.Write(buf)
			if err != nil {
				return err
			}
			return nil
		})
	})
	if err != nil {
		panic(err)
	}

	return &FuncResult{
		OutFile: fn,
	}
}
