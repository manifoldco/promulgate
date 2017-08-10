package main

import (
	"os"

	"github.com/manifoldco/promulgate/cmd"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "Make releases widely known"
	app.Commands = cmd.Cmds
	app.Run(os.Args)
}
