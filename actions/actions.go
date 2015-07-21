package actions

import (
	"encoding/json"
	"fmt"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/composer"
	"github.com/CenturyLinkLabs/zodiac/endpoint"
	"github.com/CenturyLinkLabs/zodiac/proxy"
)

const (
	ProxyAddress  = "localhost:61908"
	BasicDateTime = "2006-01-02 15:04:05"
)

var (
	DefaultComposer composer.Composer
	endpointFactory endpoint.EndpointFactory
	proxyFactory    proxy.ProxyFactory
)

func init() {
	DefaultComposer = composer.NewExecComposer(ProxyAddress)
	endpointFactory = endpoint.NewEndpoint
	proxyFactory = proxy.NewHTTPProxy
}

type Options struct {
	Args            []string
	Flags           map[string]string
	EndpointOptions endpoint.EndpointOptions
}

type Zodiaction func(Options) (prettycli.Output, error)

type DeploymentManifests []DeploymentManifest

type DeploymentManifest struct {
	Services   []Service
	DeployedAt string
	Message    string
}

type Service struct {
	Name            string
	ContainerConfig endpoint.ContainerConfig
}

func collectRequests(options Options, noBuild bool) ([]proxy.ContainerRequest, error) {
	endpoint, err := endpointFactory(options.EndpointOptions)
	if err != nil {
		return nil, err
	}
	ep := endpoint

	p := proxyFactory(ProxyAddress, ep, noBuild)

	go p.Serve()
	defer p.Stop()

	if err := DefaultComposer.Run(options.Flags); err != nil {
		return nil, err
	}

	return p.GetRequests()
}

func startServices(services []Service, manifests DeploymentManifests, endpoint endpoint.Endpoint) error {
	manifestsBlob, err := json.Marshal(manifests)
	if err != nil {
		return err
	}

	for _, svc := range services {
		if svc.ContainerConfig.Labels == nil {
			svc.ContainerConfig.Labels = make(map[string]string)
		}
		svc.ContainerConfig.Labels["zodiacManifest"] = string(manifestsBlob)

		fmt.Printf("Creating %s\n", svc.Name)

		if err := endpoint.StartContainer(svc.Name, svc.ContainerConfig); err != nil {
			return err
		}
	}

	return nil
}

func getDeploymentManifests(reqs []proxy.ContainerRequest, endpoint endpoint.Endpoint) (DeploymentManifests, error) {
	var manifests DeploymentManifests
	var inspectError error

	for _, req := range reqs {
		ci, err := endpoint.InspectContainer(req.Name)
		inspectError = err
		if err != nil {
			continue
		}

		if err := json.Unmarshal([]byte(ci.Config.Labels["zodiacManifest"]), &manifests); err != nil {
			return nil, err
		}

		break
	}

	if inspectError != nil {
		return nil, inspectError
	}

	return manifests, nil
}
