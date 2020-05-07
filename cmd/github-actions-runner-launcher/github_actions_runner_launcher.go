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
const Version = "0.0.1"

func serve(cliCtx *cli.Context) error {
	launcher, err := githubactionsrunnerlauncher.NewWithConfig(cliCtx.String("config"))
	if err != nil {
		return fmt.Errorf("Failed to create new launcher: %v", err)
	}
	launcher.Run()
	return nil
}

func main() {
	app := &cli.App{
		Name:  "github-actions-runner-launcher",
		Usage: "",
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
				Value:   "2.169.1",
				EnvVars: []string{"RUNNER_VERSION"},
				Usage:   "runner version",
			},
			&flags.LogLevelFlag,
		},
		Action: func(ctx *cli.Context) error {
			if level, err := log.ParseLevel(ctx.String("log")); err == nil {
				log.SetLevel(level)
			}
			err := serve(ctx)
			return err
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
