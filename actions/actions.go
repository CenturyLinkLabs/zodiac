package actions

import (
	"encoding/json"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/composer"
	"github.com/CenturyLinkLabs/zodiac/proxy"
	"github.com/samalba/dockerclient"
)

const (
	ProxyAddress  = "localhost:31981"
	BasicDateTime = "2006-01-02 15:04:05"
)

var (
	DefaultProxy    proxy.Proxy
	DefaultComposer composer.Composer
)

func init() {
	DefaultProxy = proxy.NewHTTPProxy(ProxyAddress)
	DefaultComposer = composer.NewExecComposer(ProxyAddress)
}

type Options struct {
	Args  []string
	Flags map[string]string
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
	ContainerConfig dockerclient.ContainerConfig
}

func collectRequests(options Options) ([]proxy.ContainerRequest, error) {
	go DefaultProxy.Serve()
	defer DefaultProxy.Stop()

	if err := DefaultComposer.Run(options.Flags); err != nil {
		return nil, err
	}

	return DefaultProxy.GetRequests()
}

func startServices(services []Service, manifests DeploymentManifests, endpoint Endpoint) error {
	manifestsBlob, err := json.Marshal(manifests)
	if err != nil {
		return err
	}

	for _, svc := range services {
		if svc.ContainerConfig.Labels == nil {
			svc.ContainerConfig.Labels = make(map[string]string)
		}
		svc.ContainerConfig.Labels["zodiacManifest"] = string(manifestsBlob)

		if err := endpoint.StartContainer(svc.Name, svc.ContainerConfig); err != nil {
			return err
		}
	}

	return nil
}

func getDeploymentManifests(reqs []proxy.ContainerRequest, endpoint Endpoint) (DeploymentManifests, error) {
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
