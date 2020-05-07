package main

import (
	"fmt"
	"os"

	"github.com/romnnn/flags4urfavecli/flags"
	"github.com/romnnn/flags4urfavecli/values"
	"github.com/romnnn/github_actions_runner_launcher"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Rev is set on build time to the git HEAD
var Rev = ""

// Version is incremented using bump2version
const Version = "0.0.1"

func serve(c *cli.Context) error {
	greeting := fmt.Sprintf("Hi %s", c.String("name"))
	log.Info(github_actions_runner_launcher.Shout(greeting))
	return nil
}

func main() {
	app := &cli.App{
		Name:  "github_actions_runner_launcher",
		Usage: "",
		Flags: []cli.Flag{
			&cli.GenericFlag{
				Name: "format",
				Value: &values.EnumValue{
					Enum:    []string{"json", "xml", "csv"},
					Default: "xml",
				},
				EnvVars: []string{"FILEFORMAT"},
				Usage:   "input file format",
			},
			&flags.LogLevelFlag,
		},
		Action: func(ctx *cli.Context) error {
			if level, err := log.ParseLevel(ctx.String("log")); err == nil {
				log.SetLevel(level)
			}
			log.Infof("Format is: %s", ctx.String("format"))
			err := serve(c)
			return err
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
