package subcommands

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// GlobalCommands global commands
var GlobalCommands = []*cli.Command{
	pipelineCmd,
	webCmd,
}

// IsValidCommand valid command name
func IsValidCommand(cmd string) bool {
	if len(cmd) == 0 {
		return false
	}
	if cmd[0] == '-' {
		switch cmd {
		case "--help", "-help", "-h", "--version", "-version", "-v":
			// 自带的配置
			return true
		default:
			for _, option := range GlobalOptions {
				for _, name := range option.Names() {
					if cmd == "--"+name || cmd == "-"+name {
						return true
					}
				}
			}
		}
		return false
	}

	for _, command := range GlobalCommands {
		if command.Name == cmd {
			return true
		}
	}
	return false
}

// GlobalOptions global options
var GlobalOptions = []cli.Flag{
	&cli.BoolFlag{
		Name:  "verbose",
		Usage: "print more information",
	},
}

// BeforAction generate fofa client
func BeforAction(context *cli.Context) error {

	// not any command, and no query for default command(search)
	if len(os.Args) == 1 {
		return nil
	}

	//logrus.SetOutput(os.Stderr) // 默认
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	if context.Bool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	return nil
}
