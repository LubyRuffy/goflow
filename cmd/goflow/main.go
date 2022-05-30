package main

import (
	"fmt"
	"github.com/LubyRuffy/goflow/cmd/goflow/subcommands"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown" // goreleaser fill

	defaultCommand = "pipeline"
)

func main() {
	app := &cli.App{
		Name:                   "goflow",
		Usage:                  fmt.Sprintf("goflow %s, commit %s, built at %s", version, commit, date),
		Version:                version,
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		Authors: []*cli.Author{
			{
				Name:  "lubyruffy",
				Email: "lubyruffy@gmail.com",
			},
		},
		Flags:    subcommands.GlobalOptions,
		Before:   subcommands.BeforAction,
		Commands: subcommands.GlobalCommands,
	}

	// default command
	if len(os.Args) > 1 && !subcommands.IsValidCommand(os.Args[1]) {
		var newArgs []string
		newArgs = append(newArgs, os.Args[0])
		newArgs = append(newArgs, defaultCommand)
		newArgs = append(newArgs, os.Args[1:]...)
		os.Args = newArgs
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
