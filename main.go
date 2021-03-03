package main

import (
	"log"
	"os"
	"strings"

	"github.com/leonelquinteros/gotext"

	"github.com/deanstalker/goas/internal/util"

	"github.com/urfave/cli"
)

var version = "v1.0.0"

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "module-path",
		Value: "",
		Usage: gotext.Get("usage.module-path"),
	},
	cli.StringFlag{
		Name:  "main-file-path",
		Value: "",
		Usage: gotext.Get("usage.main-file-path"),
	},
	cli.StringFlag{
		Name:  "handler-path",
		Value: "",
		Usage: gotext.Get("usage.handler-path"),
	},
	cli.StringFlag{
		Name:  "output",
		Value: "",
		Usage: gotext.Get("usage.output"),
	},
	cli.StringFlag{
		Name:  "format",
		Value: "json",
		Usage: gotext.Get("usage.format"),
	},
	cli.StringFlag{
		Name:  "exclude-packages",
		Value: "",
		Usage: gotext.Get("usage.exclude-packages"),
	},
	cli.BoolFlag{
		Name:  "debug",
		Usage: gotext.Get("usage.debug"),
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
	gotext.Configure(
		"./locales",
		"en",
		"default",
	)
	app := cli.NewApp()
	app.Name = gotext.Get("app.name")
	app.Usage = ""
	app.Version = version
	app.Copyright = gotext.Get("app.copyright")
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
