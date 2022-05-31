package subcommands

import (
	"errors"
	"fmt"
	"github.com/LubyRuffy/goflow"
	"github.com/LubyRuffy/goflow/gocodefuncs"
	"github.com/LubyRuffy/goflow/translater"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/lubyruffy/gofofa"
	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
)

var (
	pipelineFile    string
	pipelineTaskOut string // 导出任务列表文件
	listWorkflows   bool
)

// pipeline subcommand
var pipelineCmd = &cli.Command{
	Name:                   "pipeline",
	Usage:                  "exec workflows",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "file",
			Aliases:     []string{"f"},
			Usage:       "load pipeline file",
			Destination: &pipelineFile,
		},
		&cli.StringFlag{
			Name:        "taskOut",
			Aliases:     []string{"t"},
			Usage:       "output pipeline tasks",
			Destination: &pipelineTaskOut,
		},
		&cli.BoolFlag{
			Name:        "list",
			Aliases:     []string{"l"},
			Usage:       "list support workflows",
			Destination: &listWorkflows,
		},
	},
	Action: pipelineAction,
}

func pipelineAction(ctx *cli.Context) error {
	var err error

	if listWorkflows {
		fmt.Println(translater.Translators)
		return nil
	}

	// valid same config
	var pipelineContent string
	if len(pipelineFile) > 0 {
		v, err := os.ReadFile(pipelineFile)
		if err != nil {
			return err
		}
		pipelineContent = string(v)
	}
	if v := ctx.Args().First(); len(v) > 0 {
		if len(pipelineContent) > 0 {
			return errors.New("file and content only one is allowed")
		}
		pipelineContent, err = workflowast.NewParser().Parse(v)
		if err != nil {
			return err
		}
	}

	fofaCli, err := gofofa.NewClient()
	if err != nil {
		panic(fmt.Errorf("fofa connect err: %w", err))
	}

	pr := goflow.New().WithObject(gocodefuncs.FofaObjectName, fofaCli)
	_, err = pr.Run(pipelineContent)
	if err != nil {
		panic(err)
	}

	err = utils.EachLine(pr.LastFile, func(line string) error {
		fmt.Println(line)
		return nil
	})
	if err != nil {
		panic(err)
	}

	if len(pipelineTaskOut) > 0 {
		err = ioutil.WriteFile(pipelineTaskOut, []byte(pr.DumpTasks(false, "")), 0666)
		if err != nil {
			panic(err)
		}

		browser.OpenFile(pipelineTaskOut)
	}

	return nil
}
