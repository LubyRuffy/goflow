package gocodefuncs

import (
	"encoding/json"
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/itchyny/gojq"
	"github.com/mitchellh/mapstructure"
	"os"
)

type jqQueryParams struct {
	Query string `json:"query"`
}

// JqQuery jq command
func JqQuery(p Runner, params map[string]interface{}) *FuncResult {
	var fn string
	var err error
	var options zqQueryParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(err)
	}

	fn, err = utils.WriteTempFile(".json", nil)
	if err != nil {
		panic(err)
	}

	query, err := gojq.Parse(options.Query)
	if err != nil {
		panic(fmt.Errorf("JqQuery error: %w", err))
	}
	f, err := os.Open(p.GetLastFile())
	if err != nil {
		panic(fmt.Errorf("JqQuery error: %w", err))
	}
	defer f.Close()

	var input map[string]interface{}
	err = json.NewDecoder(f).Decode(&input)
	if err != nil {
		panic(fmt.Errorf("JqQuery error: %w", err))
	}

	fn, err = utils.WriteTempFile(".json", func(f *os.File) error {
		iter := query.Run(input) // or query.RunWithContext
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				panic(fmt.Errorf("JqQuery error: %w", err))
			}

			b, err := json.Marshal(v)
			if err != nil {
				panic(fmt.Errorf("JqQuery error: %w", err))
			}

			_, err = f.Write(b)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return &FuncResult{
		OutFile: fn,
	}
}
