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
	// ProxyAddress refers to the fake Proxy we redirect requests to.
	ProxyAddress = "localhost:61908"
	// BasicDateTime is our revolutionary time format.
	BasicDateTime = "2006-01-02 15:04:05"
)

var (
	// DefaultComposer is the composer to be used by this package. It must satisfy the Composer interface.
	DefaultComposer composer.Composer
	endpointFactory endpoint.EndpointFactory
	proxyFactory    proxy.ProxyFactory
)

func init() {
	DefaultComposer = composer.NewExecComposer(ProxyAddress)
	endpointFactory = endpoint.NewEndpoint
	proxyFactory = proxy.NewHTTPProxy
}

// Options represent the arguments for a Zodiaction
type Options struct {
	// Args supplied by the user.
	Args []string
	// Flags supplied by the user.
	Flags map[string]string
	// EndpointOptions have a comment.
	EndpointOptions endpoint.EndpointOptions
}

// a Zodiaction is wraps all the magic.
type Zodiaction func(Options) (prettycli.Output, error)

// DeploymentManifests are a collection of DeploymentManifest.
type DeploymentManifests []DeploymentManifest

// a DeploymentManifest tells us what we need to know about our destiny.
type DeploymentManifest struct {
	Services   []Service
	DeployedAt string
	Message    string
}

type Service struct {
	OriginalImage   string
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
		svc.ContainerConfig.Labels["com.centurylinklabs.zodiac.original-image"] = svc.OriginalImage

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
