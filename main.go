package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CenturyLinkLabs/prettycli"
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

type zodiaction func(c cluster.Cluster) (prettycli.Output, error)

func createHandler(z zodiaction) func(c *cli.Context) {
	return func(c *cli.Context) {
		o, err := z(endpoints)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(o.ToPrettyOutput())
	}
}
