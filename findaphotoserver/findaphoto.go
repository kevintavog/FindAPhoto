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
	app.Spec = "[-d] [-i]"
	developmentMode := app.BoolOpt("d", false, "Development mode (hit <enter> to exit, listen on a different port, use a different index)")
	indexOverride := app.StringOpt("i", "", "The index to use")

	app.Action = func() { run(*developmentMode, *indexOverride) }

	app.Run(os.Args)
}
