package main

import (
	"context"
	"corvina/corvina-seed/src/cmd"
	"corvina/corvina-seed/src/utils"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	cli "github.com/urfave/cli/v2"
)

var verboseFlag *cli.BoolFlag = &cli.BoolFlag{
	Name:    "verbose",
	Aliases: []string{"v"},
	Usage:   "Enable verbose mode",
}

var keycloakFlags []*cli.StringFlag = []*cli.StringFlag{
	{
		Name:        "keycloak-origin",
		Aliases:     []string{"ko"},
		Usage:       "Keycloak origin",
		DefaultText: "https://auth.corvina.fog:10443",
		Value:       "https://auth.corvina.fog:10443",
	},
	{
		Name:        "keycloak-master-user",
		Aliases:     []string{"ku"},
		Usage:       "Keycloak master user",
		DefaultText: "keycloak-admin",
		Value:       "keycloak-admin",
	},
	{
		Name:        "keycloak-master-pass",
		Aliases:     []string{"kp"},
		Usage:       "Keycloak master password",
		DefaultText: "keycloak-admin",
		Value:       "keycloak-admin",
	},
}

var adminUserFlag *cli.StringFlag = &cli.StringFlag{
	Name:        "admin-user",
	Aliases:     []string{"au"},
	Usage:       "Admin user",
	DefaultText: "admin@exor",
	Value:       "admin@exor",
}

var originFlag *cli.StringFlag = &cli.StringFlag{
	Name:        "origin",
	Aliases:     []string{"o"},
	Usage:       "Corvina origin",
	DefaultText: "https://app.corvina.fog:10443",
	Value:       "https://app.corvina.fog:10443",
}

func getDomainFromOrigin(origin string) string {
	return strings.Replace(origin, "https://app.", "", 1)
}

func getUserRealmFromAdminUser(adminUser string) string {
	return strings.Split(adminUser, "@")[1]
}

func main() {
	utils.InitLog()

	app := &cli.App{
		Name:  "corvina-seed",
		Usage: "make an explosive entrance",
		Action: func(c *cli.Context) error {
			cli.ShowAppHelpAndExit(c, 0)
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

				c.Context = context.WithValue(c.Context, utils.OriginKey, c.String("origin"))
				domain := getDomainFromOrigin(c.String("origin"))
				c.Context = context.WithValue(c.Context, utils.DomainKey, domain)
				c.Context = context.WithValue(c.Context, utils.LicenseHostKey, "https://app."+domain+"/svc/license")
				c.Context = context.WithValue(c.Context, utils.LicenseManagerUser, c.String("license-manager-user"))
				c.Context = context.WithValue(c.Context, utils.LicenseManagerPass, c.String("license-manager-pass"))
				c.Context = context.WithValue(c.Context, utils.KeycloakOrigin, c.String("keycloak-origin"))
				c.Context = context.WithValue(c.Context, utils.AdminUserKey, c.String("admin-user"))
				c.Context = context.WithValue(c.Context, utils.UserRealm, getUserRealmFromAdminUser(c.String("admin-user")))
				c.Context = context.WithValue(c.Context, utils.KeycloakMasterUser, c.String("keycloak-master-user"))
				c.Context = context.WithValue(c.Context, utils.KeycloakMasterPass, c.String("keycloak-master-pass"))
				c.Context = context.WithValue(c.Context, utils.DeviceCount, c.Int64("device-count"))
				c.Context = context.WithValue(c.Context, utils.EachDeviceHasMapping, c.Bool("each-device-has-mapping"))
				c.Context = context.WithValue(c.Context, utils.DeviceGroupCount, c.Int64("device-group-count"))
				c.Context = context.WithValue(c.Context, utils.ModelCount, c.Int64("model-count"))

				return cmd.Run(c.Context)
			},
			Flags: []cli.Flag{
				verboseFlag,
				originFlag,
				keycloakFlags[0],
				keycloakFlags[1],
				keycloakFlags[2],
				adminUserFlag,
				&cli.Int64Flag{
					Name:    "model-count",
					Aliases: []string{"m"},
					Usage:   "Number of models to create",
				},
				&cli.Int64Flag{
					Name:    "device-count",
					Aliases: []string{"d"},
					Usage:   "Number of devices to create (automatically creates device license)",
				},
				&cli.BoolFlag{
					Name:        "each-device-has-mapping",
					Usage:       "If true, this cli will create a model/mapping for each device created when --device-count is provided",
					Aliases:     []string{"edm"},
					DefaultText: "true",
					Value:       true,
				},
				&cli.Int64Flag{
					Name:    "device-group-count",
					Aliases: []string{"dg"},
					Usage:   "Number of device groups to create",
				},
				&cli.StringFlag{
					Name:        "license-manager-user",
					Aliases:     []string{"lmu"},
					Usage:       "License manager user",
					DefaultText: "license-admin",
					Value:       "license-admin",
				},
				&cli.StringFlag{
					Name:        "license-manager-pass",
					Aliases:     []string{"lmp"},
					Usage:       "License manager password",
					DefaultText: "license-admin",
					Value:       "license-admin",
				},
			},
		},
		{
			Name:  "master-token",
			Usage: "Get the master token for the keycloak master user",
			Action: func(c *cli.Context) error {
				if c.Bool("verbose") {
					utils.VerboseLog()
				}

				c.Context = context.WithValue(c.Context, utils.KeycloakOrigin, c.String("keycloak-origin"))
				c.Context = context.WithValue(c.Context, utils.KeycloakMasterUser, c.String("keycloak-master-user"))
				c.Context = context.WithValue(c.Context, utils.KeycloakMasterPass, c.String("keycloak-master-pass"))

				return cmd.MasterToken(c.Context)
			},
			Flags: []cli.Flag{
				verboseFlag,
				keycloakFlags[0],
				keycloakFlags[1],
				keycloakFlags[2],
			},
		},
		{
			Name:  "admin-token",
			Usage: "Get the token for the provided admin user",
			Action: func(c *cli.Context) error {
				if c.Bool("verbose") {
					utils.VerboseLog()
				}

				c.Context = context.WithValue(c.Context, utils.OriginKey, c.String("origin"))
				domain := getDomainFromOrigin(c.String("origin"))
				c.Context = context.WithValue(c.Context, utils.DomainKey, domain)
				c.Context = context.WithValue(c.Context, utils.KeycloakOrigin, c.String("keycloak-origin"))
				c.Context = context.WithValue(c.Context, utils.KeycloakMasterUser, c.String("keycloak-master-user"))
				c.Context = context.WithValue(c.Context, utils.KeycloakMasterPass, c.String("keycloak-master-pass"))
				c.Context = context.WithValue(c.Context, utils.AdminUserKey, c.String("admin-user"))
				c.Context = context.WithValue(c.Context, utils.UserRealm, getUserRealmFromAdminUser(c.String("admin-user")))

				return cmd.AdminToken(c.Context)
			},
			Flags: []cli.Flag{
				verboseFlag,
				originFlag,
				keycloakFlags[0],
				keycloakFlags[1],
				keycloakFlags[2],
				adminUserFlag,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("Error running the app")
	}
}
