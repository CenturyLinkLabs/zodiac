package actions

import (
	"encoding/json"
	"fmt"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/cluster"
	"github.com/CenturyLinkLabs/zodiac/composer"
	"github.com/CenturyLinkLabs/zodiac/proxy"
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

const ProxyAddress = "localhost:31981"

var (
	DefaultProxy    = proxy.NewHTTPProxy(ProxyAddress)
	DefaultComposer = composer.NewExecComposer(ProxyAddress)
)

func Deploy(c cluster.Cluster, args []string) (prettycli.Output, error) {
	for _, endpoint := range c.Endpoints() {
		// TODO: handle error
		go DefaultProxy.Serve(endpoint)
		// TODO: handle error
		defer DefaultProxy.Stop()

		// TODO: handle error
		// TODO: args not passed!
		DefaultComposer.Run(args)

		reqs := DefaultProxy.DrainRequests()

		for _, req := range reqs {
			client, err := dockerclient.NewDockerClient(endpoint.Name(), nil)
			newManifest := req.CreateOptions
			fmt.Println(req.Name)

			var cc dockerclient.ContainerConfig
			if err := json.Unmarshal(req.CreateOptions, &cc); err != nil {
				log.Fatalf("error unmarshalling request JSON for '%s': %s", req.Name, err.Error())
			}

			if cc.Labels == nil {
				cc.Labels = make(map[string]string)
			}
			cc.Labels["zodiacManifest"] = string(newManifest)

			id, err := client.CreateContainer(&cc, req.Name)
			if err != nil {
				log.Fatalf("Problem creating container: ", err)
			}

			log.Infof("%s created as %s", req.Name, id)

			if err := client.StartContainer(id, &dockerclient.HostConfig{}); err != nil {
				log.Fatal("problem starting: ", err)
			}
		}
		// Phase Deux: Build current manifest

		// Phase Deux: Fetch current deployments

		// Phase Deux: Build new Manifests from ContainerRequest + Old Manifest

		// Phase Un: (Pull?)+Create+Start Containers on all hosts
		// TODO: handle error
		// c.StartContainers(DefaultComposer.DrainRequests())

		// Phase Deux: ^ injecting manifest before Create
	}

	return prettycli.PlainOutput{"whatevs"}, nil
}
