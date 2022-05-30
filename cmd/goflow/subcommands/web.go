package subcommands

import (
	"github.com/LubyRuffy/goflow/web"
	"github.com/urfave/cli/v2"
)

var (
	listenAddr string
)

// web subcommand
var webCmd = &cli.Command{
	Name:  "web",
	Usage: "fofa web interface",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "addr",
			Value:       ":5555",
			Usage:       "web listen addr",
			Destination: &listenAddr,
		},
	},
	Action: func(ctx *cli.Context) error {
		return web.Start(listenAddr)
	},
}
