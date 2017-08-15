package main

import (
	"os"

	"github.com/manifoldco/promulgate/cmd"
	"github.com/urfave/cli"
)

const version = "dev"

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Usage = "Make releases widely known"
	app.Commands = cmd.Cmds
	app.Run(os.Args)
}
