package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/brimdata/zed/cli/zq"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

type zqValueParams struct {
	Query string `json:"query"`
}

// ZqValue zq command计算值返回
func ZqValue(p Runner, params map[string]interface{}) string {
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

	cmd := []string{"-f", "text", "-o", fn, options.Query, p.GetLastFile()}
	logrus.Debugf("zq cmd: %v", cmd)
	err = zq.Cmd.ExecRoot(cmd)
	if err != nil {
		panic(fmt.Errorf("ZqValue error: %w", err))
	}

	d, err := utils.ReadFirstLineOfFile(fn)
	if err != nil {
		panic(fmt.Errorf("ZqValue error: %w", err))
	}

	return string(d)
}
