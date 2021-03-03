package main

import (
	"log"
	"os"
	"strings"

	"github.com/deanstalker/goas/internal/util"

	"github.com/urfave/cli"
)

var version = "v1.0.0"

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "module-path",
		Value: "",
		Usage: "goas will search @comment under the module",
	},
	cli.StringFlag{
		Name:  "main-file-path",
		Value: "",
		Usage: "goas will start to search @comment from this main file",
	},
	cli.StringFlag{
		Name:  "handler-path",
		Value: "",
		Usage: "goas only search handleFunc comments under the path",
	},
	cli.StringFlag{
		Name:  "output",
		Value: "",
		Usage: "output file",
	},
	cli.StringFlag{
		Name:  "format",
		Value: "json",
		Usage: "json (default) or yaml format - for stdout only",
	},
	cli.StringFlag{
		Name:  "exclude-packages",
		Value: "",
		Usage: "Exclude by package name eg. integration",
	},
	cli.BoolFlag{
		Name:  "debug",
		Usage: "show debug message",
	},
}

func action(c *cli.Context) error {
	p, err := newParser(
		util.ModulePath(c.GlobalString("module-path")),
		c.GlobalString("main-file-path"),
		c.GlobalString("handler-path"),
		c.GlobalString("exclude-packages"),
		c.GlobalBool("debug"),
	)
	if err != nil {
		return err
	}

	output := util.CLIOutput(c.GlobalString("output"))
	format := c.GlobalString("format")

	outputFormat := output.GetFormat()
	if format != "" {
		outputFormat = strings.ToLower(format)
	}
	_, err = p.CreateOAS(c.GlobalString("output"), output.GetMode(), outputFormat)
	return err
}

func main() {
	app := cli.NewApp()
	app.Name = "goas"
	app.Usage = ""
	app.Version = version
	app.Copyright = "(c) 2018 mikun800527@gmail.com"
	app.HideHelp = true
	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		_ = cli.ShowAppHelp(c)
		return nil
	}
	app.Flags = flags
	app.Action = action

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
