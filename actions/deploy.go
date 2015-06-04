package actions

import (
	"encoding/json"
	"fmt"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/cluster"
	"github.com/CenturyLinkLabs/zodiac/composer"
	"github.com/CenturyLinkLabs/zodiac/proxy"
	"github.com/samalba/dockerclient"
)

const ProxyAddress = "localhost:31981"

var (
	DefaultProxy    proxy.Proxy
	DefaultComposer composer.Composer
)

func init() {
	DefaultProxy = proxy.NewHTTPProxy(ProxyAddress)
	DefaultComposer = composer.NewExecComposer(ProxyAddress)
}

func Deploy(c cluster.Cluster, args []string) (prettycli.Output, error) {

	var reqs []cluster.ContainerRequest

	for _, endpoint := range c.Endpoints() {

		//client, _ := dockerclient.NewDockerClient(endpoint.Name(), nil)
		// TODO: handle error
		go DefaultProxy.Serve(endpoint)
		// TODO: handle error
		defer DefaultProxy.Stop()

		// TODO: handle error
		// TODO: args not passed!
		DefaultComposer.Run(args)
		reqs = DefaultProxy.DrainRequests()

		for _, req := range reqs {

			//fmt.Println(req.Name)

			var cc dockerclient.ContainerConfig

			if err := json.Unmarshal(req.CreateOptions, &cc); err != nil {
				return nil, err
			}

			imageId, err := endpoint.ResolveImage(cc.Image)
			if err != nil {
				return nil, err
			}

			cc.Image = imageId

			newManifest, err := json.Marshal(cc)
			if err != nil {
				return nil, err
			}

			if cc.Labels == nil {
				cc.Labels = make(map[string]string)
			}
			cc.Labels["zodiacManifest"] = string(newManifest)
			// Phase Deux: Build current manifest

			// Phase Deux: Fetch current deployments

			// Phase Deux: Build new Manifests from ContainerRequest + Old Manifest

			// Phase Un: (Pull?)+Create+Start Containers on all hosts
			// TODO: handle error

			//endpoint.StartContainers(DefaultProxy.DrainRequests())
			endpoint.StartContainer(req.Name, cc)
			// Phase Deux: ^ injecting manifest before Create
		}
	}

	output := fmt.Sprintf("Successfully deployed %d container(s)", len(reqs))
	return prettycli.PlainOutput{output}, nil
}
