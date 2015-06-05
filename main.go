package main // import "github.com/CenturyLinkLabs/zodiac"

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CenturyLinkLabs/zodiac/actions"
	"github.com/CenturyLinkLabs/zodiac/discovery"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const version = "0.0.1"

var (
	commands []cli.Command
	cluster  discovery.HardcodedCluster
)

func init() {
	log.SetLevel(log.WarnLevel)

	commands = []cli.Command{
		{
			Name:   "verify",
			Usage:  "Verify the cluster",
			Action: verifyAction,
			Before: requireCluster,
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
	endpoints := c.GlobalString("cluster")
	if endpoints == "" {
		err := errors.New("you must specify a cluster to connect to")
		log.Error(err)
		return err
	}

	cluster = discovery.HardcodedCluster{}
	for _, s := range strings.Split(endpoints, ",") {
		cluster = append(cluster, &discovery.DockerEndpoint{URL: s})
	}

	return nil
}

func verifyAction(c *cli.Context) {
	o, err := actions.Verify(cluster)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(o.ToPrettyOutput())
}
