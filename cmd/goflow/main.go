package main

import (
	"flag"
	"fmt"
	"github.com/LubyRuffy/goflow"
	"github.com/LubyRuffy/goflow/internal/utils"
	"github.com/LubyRuffy/goflow/translater"
	"github.com/LubyRuffy/goflow/workflowast"
	"github.com/lubyruffy/gofofa"
	"github.com/pkg/browser"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	listWorkflows := flag.Bool("list", false, "list support workflows")
	pipelineTaskOut := flag.String("taskOut", "", "output pipeline tasks")
	flag.Parse()

	var err error
	var fofaCli *gofofa.Client

	if *listWorkflows {
		fmt.Println(translater.Translators)
		return
	}

	if len(flag.Args()) == 0 {
		log.Println("no flow to execute, usage: <workflow sentences>")
		return
	}

	fofaCli, err = gofofa.NewClient()
	if err != nil {
		panic(fmt.Errorf("fofa connect err: %w", err))
	}

	pipelineContent := strings.Join(flag.Args(), " ")
	gocode, err := workflowast.NewParser().Parse(pipelineContent)
	if err != nil {
		panic(fmt.Errorf("fofa connect err: %w", err))
	}

	pr := goflow.New(goflow.WithObject("fofacli", fofaCli))
	_, err = pr.Run(gocode)
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

	if len(*pipelineTaskOut) > 0 {
		err = ioutil.WriteFile(*pipelineTaskOut, []byte(pr.DumpTasks(false)), 0666)
		if err != nil {
			panic(err)
		}

		browser.OpenFile(*pipelineTaskOut)
	}
}
