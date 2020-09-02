package main

import (
	"fmt"
	"os"

	"github.com/romnnn/flags4urfavecli/flags"
	"github.com/romnnn/flags4urfavecli/values"
	githubactionsrunnerlauncher "github.com/romnnn/github-actions-runner-launcher"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Rev is set on build time to the git HEAD
var Rev = ""

// Version is incremented using bump2version
const Version = "0.1.2"

func serve(cliCtx *cli.Context, run bool) error {
	launcher, err := githubactionsrunnerlauncher.NewWithConfig(cliCtx.String("config"))
	launcher.RemoveExisting = cliCtx.Bool("remove")
	launcher.Reconfigure = cliCtx.Bool("reconfigure")
	if err != nil {
		return fmt.Errorf("Failed to create new launcher: %v", err)
	}
	return launcher.Run(run)
}

func main() {
	app := &cli.App{
		Name:  "github-actions-runner-launcher",
		Usage: "",
		Commands: []*cli.Command{
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "install and prepare the runners",
				Action: func(ctx *cli.Context) error {
					if level, err := log.ParseLevel(ctx.String("log")); err == nil {
						log.SetLevel(level)
					}
					err := serve(ctx, false)
					return err
				},
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "start the runners",
				Action: func(ctx *cli.Context) error {
					if level, err := log.ParseLevel(ctx.String("log")); err == nil {
						log.SetLevel(level)
					}
					err := serve(ctx, true)
					return err
				},
			},
		},
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "config",
				EnvVars: []string{"CONFIG"},
				Usage:   "runner config file",
			},
			&cli.GenericFlag{
				Name: "runner-arch",
				Value: &values.EnumValue{
					Enum:    []string{"x64", "arm", "arm64"},
					Default: "x64",
				},
				EnvVars: []string{"RUNNER_ARCH"},
				Usage:   "runner host architecture",
			},
			&cli.StringFlag{
				Name:    "runner-version",
				Value:   "2.273.0",
				EnvVars: []string{"RUNNER_VERSION"},
				Usage:   "runner version",
			},
			&cli.BoolFlag{
				Name:    "remove",
				Value:   false,
				EnvVars: []string{"REMOVE"},
				Usage:   "remove any existing runners",
			},
			&cli.BoolFlag{
				Name:    "reconfigure",
				Value:   false,
				EnvVars: []string{"RECONFIGURE"},
				Usage:   "reconfigure the runners",
			},
			&flags.LogLevelFlag,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
