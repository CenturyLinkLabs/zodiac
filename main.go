package main // import "github.com/CenturyLinkLabs/zodiac"

import (
	"errors"
	"fmt"
	"os"

	"github.com/CenturyLinkLabs/zodiac/actions"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const version = "0.0.1"

var (
	commands []cli.Command
)

func init() {
	log.SetLevel(log.WarnLevel)

	commands = []cli.Command{
		{
			Name:   "verify",
			Usage:  "Verify the endpoint",
			Action: createHandler(actions.Verify),
			Before: requireCluster,
		},
		{
			Name:   "deploy",
			Usage:  "Deploy a Docker compose template",
			Action: createHandler(actions.Deploy),
			Before: requireCluster,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "message, m",
					Usage: "Give your deployment a comment",
				},
				cli.StringFlag{
					Name:   "name, n",
					Usage:  "Specify a custom project name",
					Value:  "zodiac",
					EnvVar: "ZODIAC_PROJECT_NAME",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "Specify an alternate compose file",
					Value: "docker-compose.yml",
				},
			},
		},
		{
			Name:        "rollback",
			Usage:       "rollback a deployment",
			Description: "Specify the deployment ID as the argument to the rollback command. If the deployment ID is ommitted, the most recent deployment will be assumed",
			Action:      createHandler(actions.Rollback),
			Before:      requireCluster,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "message, m",
					Usage: "Give your rollback a comment (defaults to 'Rollback to: [target deployment comment]')",
				},
				cli.StringFlag{
					Name:   "name, n",
					Usage:  "Specify a custom project name",
					Value:  "zodiac",
					EnvVar: "ZODIAC_PROJECT_NAME",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "Specify an alternate compose file",
					Value: "docker-compose.yml",
				},
			},
		},
		{
			Name:   "list",
			Usage:  "List all known deployments",
			Action: createHandler(actions.List),
			Before: requireCluster,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "name, n",
					Usage:  "Specify a custom project name",
					Value:  "zodiac",
					EnvVar: "ZODIAC_PROJECT_NAME",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "Specify an alternate compose file",
					Value: "docker-compose.yml",
				},
			},
		},
		{
			Name:   "teardown",
			Usage:  "Remove running services and deployment history for this application",
			Action: createHandler(actions.Teardown),
			Before: requireCluster,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "name, n",
					Usage:  "Specify a custom project name",
					Value:  "zodiac",
					EnvVar: "ZODIAC_PROJECT_NAME",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "Specify an alternate compose file",
					Value: "docker-compose.yml",
				},
			},
		},
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "zodiac"
	app.Version = version
	app.Usage = "Simple Docker deployment utility."
	app.Authors = []cli.Author{{"CenturyLink Labs", "clt-labs-futuretech@centurylink.com"}}
	app.Commands = commands
	app.Before = initializeCLI
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable verbose logging",
		},
		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "Docker endpoint",
			EnvVar: "ZODIAC_DOCKER_ENDPOINT",
		},
	}

	app.Run(os.Args)
}

func initializeCLI(c *cli.Context) error {
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	return nil
}

func requireCluster(c *cli.Context) error {
	arg := c.GlobalString("endpoint")
	if arg == "" {
		err := errors.New("you must specify a Docker endpoint to connect to")
		log.Error(err)
		return err
	}

	return nil
}

func createHandler(z actions.Zodiaction) func(c *cli.Context) {
	return func(c *cli.Context) {
		flags := map[string]string{}

		//TODO: is this a codegangsta bug, GlobalFlagNames?
		for _, flagName := range c.GlobalFlagNames() {
			flags[flagName] = c.String(flagName)
		}
		flags["endpoint"] = c.GlobalString("endpoint")

		actionOpts := actions.Options{
			Args:  c.Args(),
			Flags: flags,
		}

		o, err := z(actionOpts)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		fmt.Println(o.ToPrettyOutput())
	}
}
