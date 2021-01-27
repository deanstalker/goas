package main

import (
	"log"
	"os"
	"strings"

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
	cli.BoolFlag{
		Name:  "debug",
		Usage: "show debug message",
	},
}

func action(c *cli.Context) error {
	p, err := newParser(
		c.GlobalString("module-path"),
		c.GlobalString("main-file-path"),
		c.GlobalString("handler-path"),
		c.GlobalBool("debug"))
	if err != nil {
		return err
	}

	mode := ModeStdOut
	format := FormatJSON
	if c.GlobalString("output") != "" {
		if strings.Contains(c.GlobalString("output"), "json") {
			mode = ModeFileWriter
		}
		if strings.Contains(c.GlobalString("output"), "yaml") ||
			strings.Contains(c.GlobalString("output"), "yml") {
			mode = ModeFileWriter
			format = FormatYAML
		}
	}
	if c.GlobalString("format") != "" {
		format = c.GlobalString("format")
	}

	return p.CreateOAS(c.GlobalString("output"), mode, format)
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
