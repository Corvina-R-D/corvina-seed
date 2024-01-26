package main

import (
	"context"
	"corvina/corvina-seed/src/cmd"
	"corvina/corvina-seed/src/utils"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	cli "github.com/urfave/cli/v2"
)

var verboseFlag *cli.BoolFlag = &cli.BoolFlag{
	Name:    "verbose",
	Aliases: []string{"v"},
	Usage:   "Enable verbose mode",
}

func main() {
	utils.InitLog()

	app := &cli.App{
		Name:  "corvina-seed",
		Usage: "make an explosive entrance",
		Action: func(*cli.Context) error {
			fmt.Println(`List of available commands:
	- version: Print the current cli version
	- run: Start creating some entities in corvina if enough information is provided, otherwise it will start an interactive session`)
			return nil
		},
		Flags: []cli.Flag{
			verboseFlag,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:  "version",
			Usage: "Print the version",
			Action: func(c *cli.Context) error {
				cmd.Version()
				return nil
			},
		},
		{
			Name:  "run",
			Usage: "Start creating some entities in corvina if enough information is provided, otherwise it will start an interactive session",
			Action: func(c *cli.Context) error {
				if c.Bool("verbose") {
					utils.VerboseLog()
				}

				c.Context = context.WithValue(c.Context, cmd.ApiKey, c.String("api-key"))
				c.Context = context.WithValue(c.Context, cmd.DeviceCount, c.Int64("device-count"))

				return cmd.Run(c.Context)
			},
			Flags: []cli.Flag{
				verboseFlag,
				&cli.StringFlag{
					Name:    "api-key",
					Aliases: []string{"k"},
					Usage:   "Corvina API key",
				},
				&cli.Int64Flag{
					Name:    "device-count",
					Aliases: []string{"d"},
					Usage:   "Number of devices to create",
				},
				&cli.Int64Flag{
					Name:    "device-group-count",
					Aliases: []string{"dg"},
					Usage:   "Number of device groups to create",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("Error running the app")
	}
}
