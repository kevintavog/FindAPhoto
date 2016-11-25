package main

import (
	"os"

	"github.com/jawher/mow.cli"

	"github.com/kevintavog/findaphoto/common"
	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"
)

func main() {
	configuration.ReadConfiguration()
	common.ConfigureLogging(common.LogDirectory, "findaphotoserver")

	app := cli.App("findaphotoserver", "The FindAPhoto server")
	app.Spec = "[-d]"
	developmentMode := app.BoolOpt("d", false, "Development mode (hit <enter> to exit, listen on a different port, use a different index)")
	app.Action = func() { run(*developmentMode) }

	app.Run(os.Args)
}
