package main // import "github.com/CenturyLinkLabs/zodiac"

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CenturyLinkLabs/zodiac/actions"
	"github.com/CenturyLinkLabs/zodiac/cluster"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const version = "0.0.1"

var (
	commands  []cli.Command
	endpoints cluster.HardcodedCluster
)

func init() {
	log.SetLevel(log.WarnLevel)

	commands = []cli.Command{
		{
			Name:   "verify",
			Usage:  "Verify the cluster",
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
					Name:  "name, n",
					Usage: "Specify a custom project name",
					Value: "zodiac",
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
					Name:  "name, n",
					Usage: "Specify a custom project name",
					Value: "zodiac",
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
					Name:  "name, n",
					Usage: "Specify a custom project name",
					Value: "zodiac",
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
			Name:   "cluster",
			Usage:  "Use a comma-separated list of Docker endpoints",
			EnvVar: "ZODIAC_CLUSTER",
		},
	}

	app.Run(os.Args)
}

func initializeCLI(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	return nil
}

func requireCluster(c *cli.Context) error {
	arg := c.GlobalString("cluster")
	if arg == "" {
		err := errors.New("you must specify a cluster to connect to")
		log.Error(err)
		return err
	}

	endpoints = cluster.HardcodedCluster{}
	for _, host := range strings.Split(arg, ",") {
		c, err := cluster.NewDockerEndpoint(host)
		if err != nil {
			log.Fatalf("there was a problem with endpoint '%s': %s", host, err)
		}

		endpoints = append(endpoints, c)
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

		actionOpts := actions.Options{
			Args:  c.Args(),
			Flags: flags,
		}

		o, err := z(endpoints, actionOpts)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(o.ToPrettyOutput())
	}
}
