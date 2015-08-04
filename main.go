package main // import "github.com/CenturyLinkLabs/zodiac"

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CenturyLinkLabs/zodiac/actions"
	"github.com/CenturyLinkLabs/zodiac/endpoint"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const version = "0.2.0"

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
			Name:    "list",
			Aliases: []string{"history"},
			Usage:   "List all known deployments",
			Action:  createHandler(actions.List),
			Before:  requireCluster,
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
			Action: createHandlerWithConfirm(actions.Teardown, "Are you sure you want to remove the deployment and all history?"),
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
				cli.StringFlag{
					Name:  "confirm, c",
					Usage: "specify confirmation up front instead of waiting for prompt",
					Value: "y/N",
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
			EnvVar: "DOCKER_HOST",
		},
		cli.BoolTFlag{
			Name:   "tls",
			Usage:  "Use TLS",
			EnvVar: "DOCKER_TLS",
		},
		cli.BoolTFlag{
			Name:   "tlsverify",
			Usage:  "Use TLS and verify the remote",
			EnvVar: "DOCKER_TLS_VERIFY",
		},
		cli.StringFlag{
			Name:  "tlscacert",
			Usage: "Trust certs signed only by this CA",
			Value: fmt.Sprintf("%s/ca.pem", rootCertPath()),
		},
		cli.StringFlag{
			Name:  "tlscert",
			Usage: "Path to TLS certificate file",
			Value: fmt.Sprintf("%s/cert.pem", rootCertPath()),
		},
		cli.StringFlag{
			Name:  "tlskey",
			Usage: "Path to TLS key file",
			Value: fmt.Sprintf("%s/key.pem", rootCertPath()),
		},
	}

	app.Run(os.Args)
}

func rootCertPath() string {
	if os.Getenv("DOCKER_CERT_PATH") != "" {
		return os.Getenv("DOCKER_CERT_PATH")
	}
	return "~/.docker"
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

func createHandlerWithConfirm(z actions.Zodiaction, msg string) func(c *cli.Context) {
	return func(c *cli.Context) {
		cfrm := strings.ToLower(c.String("confirm"))

		if cfrm == "y" || cfrm == "yes" {
			handler(z, c)
		} else {
			fmt.Println(fmt.Sprintf("%s (y/N)", msg))
			var response string
			_, err := fmt.Scanln(&response)
			if err != nil {
				log.Fatal(err)
			}

			response = strings.ToLower(response)

			if response != "y" && response != "yes" {
				fmt.Println("Cancelled")
			} else {
				handler(z, c)
			}
		}
	}
}

func createHandler(z actions.Zodiaction) func(c *cli.Context) {
	return func(c *cli.Context) {
		handler(z, c)
	}
}

func handler(z actions.Zodiaction, c *cli.Context) {
	flags := map[string]string{}

	//TODO: is this a codegangsta bug, GlobalFlagNames?
	for _, flagName := range c.GlobalFlagNames() {
		flags[flagName] = c.String(flagName)
	}

	eOpts := endpoint.EndpointOptions{
		Host:      c.GlobalString("endpoint"),
		TLS:       c.GlobalBool("tls"),
		TLSVerify: c.GlobalBool("tlsverify"),
		TLSCaCert: c.GlobalString("tlscacert"),
		TLSCert:   c.GlobalString("tlscert"),
		TLSKey:    c.GlobalString("tlskey"),
	}

	actionOpts := actions.Options{
		Args:            c.Args(),
		Flags:           flags,
		EndpointOptions: eOpts,
	}

	o, err := z(actionOpts)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(o.ToPrettyOutput())
}
