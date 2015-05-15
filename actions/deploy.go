package actions

import (
	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/cluster"
	"github.com/CenturyLinkLabs/zodiac/composer"
	"github.com/CenturyLinkLabs/zodiac/proxy"
)

const ProxyAddress = "localhost:31981"

var (
	DefaultProxy    = proxy.NewHTTPProxy(ProxyAddress)
	DefaultComposer = composer.NewExecComposer(ProxyAddress)
)

func Deploy(c cluster.Cluster) (prettycli.Output, error) {
	// TODO: handle error
	go DefaultProxy.Serve()
	// TODO: handle error
	defer DefaultProxy.Stop()

	// TODO: handle error
	// TODO: args not passed!
	DefaultComposer.Run("bogus", "unimplemented", "args")

	// Phase Deux: Build current manifest

	// Phase Deux: Fetch current deployments

	// Phase Deux: Build new Manifests from ContainerRequest + Old Manifest

	// Phase Un: (Pull?)+Create+Start Containers on all hosts
	// TODO: handle error
	c.StartContainers(DefaultComposer.DrainRequests())

	// Phase Deux: ^ injecting manifest before Create

	return prettycli.PlainOutput{"whatevs"}, nil
}
