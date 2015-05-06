package main

import (
	"fmt"
	"os"

	"github.com/CenturyLinkLabs/zodiac/actions"
	"github.com/CenturyLinkLabs/zodiac/discovery"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const version = "0.0.1"

var commands []cli.Command

func init() {
	log.SetLevel(log.WarnLevel)

	commands = []cli.Command{
		{
			Name:   "verify",
			Usage:  "Verify the known nodes",
			Action: verifyAction,
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
	}

	app.Run(os.Args)
}

func initializeCLI(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	return nil
}

func verifyAction(c *cli.Context) {
	cluster := discovery.HardcodedCluster{{URL: "tcp://10.134.246.158:2375"}}
	o, err := actions.Verify(cluster)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(o.ToPrettyOutput())
}
