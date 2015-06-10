package actions

import (
	"encoding/json"
	"fmt"
	"time"

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

type DeploymentManifests []DeploymentManifest

type DeploymentManifest struct {
	Services   []Service
	DeployedAt string
}

type Service struct {
	Name            string
	ContainerConfig dockerclient.ContainerConfig
}

func Rollback(c cluster.Cluster, args []string) (prettycli.Output, error) {
	var reqs []cluster.ContainerRequest

	for _, endpoint := range c.Endpoints() {

		reqs = collectRequests(endpoint, args)
		// Get most recent deployment's manifests
		var manifests DeploymentManifests
		for _, req := range reqs {
			ci, err := endpoint.InspectContainer(req.Name)
			if err != nil {
				continue
			}

			if err := json.Unmarshal([]byte(ci.Config.Labels["zodiacManifest"]), &manifests); err != nil {
				fmt.Printf("ERROR: %s\n\n", err)
				return nil, err
			}

			break
		}

		// shut down current deployment

		currentDeployment := manifests[len(manifests)-1]

		for _, svc := range currentDeployment.Services {
			endpoint.RemoveContainer(svc.Name)
		}

		// TODO: allow passing in index
		newDeployment := manifests[len(manifests)-2]

		manifests = append(manifests, newDeployment)
		newDeployment = manifests[len(manifests)-1]
		manifests[len(manifests)-1].DeployedAt = time.Now().Format(time.RFC3339)

		startServices(newDeployment.Services, manifests, endpoint)
	}

	output := fmt.Sprintf("Successfully rolled back to %d container(s)", len(reqs))
	return prettycli.PlainOutput{output}, nil
}

func Deploy(c cluster.Cluster, args []string) (prettycli.Output, error) {

	var reqs []cluster.ContainerRequest

	for _, endpoint := range c.Endpoints() {

		reqs = collectRequests(endpoint, args)

		dm := DeploymentManifest{
			Services:   []Service{},
			DeployedAt: time.Now().Format(time.RFC3339),
		}

		for _, req := range reqs {
			s, err := serviceForRequest(req)
			if err != nil {
				return nil, err
			}

			imageId, err := endpoint.ResolveImage(s.ContainerConfig.Image)
			if err != nil {
				return nil, err
			}

			s.ContainerConfig.Image = imageId

			dm.Services = append(dm.Services, s)
		}

		oldManifestBlob := "[]"
		for _, svc := range dm.Services {
			ci, err := endpoint.InspectContainer(svc.Name)

			if err == nil {
				err := endpoint.RemoveContainer(svc.Name)
				if err != nil {
					//TODO: figure out if we really want to abort here
					return nil, err
				}
			}

			// TODO, only assign if not empty
			if (ci != nil) && (ci.Config != nil) && (ci.Config.Labels != nil) && (ci.Config.Labels["zodiacManifest"] != "") {
				oldManifestBlob = ci.Config.Labels["zodiacManifest"]
			}
		}

		var manifests DeploymentManifests
		if err := json.Unmarshal([]byte(oldManifestBlob), &manifests); err != nil {
			return nil, err
		}
		manifests = append(manifests, dm)

		startServices(dm.Services, manifests, endpoint)
	}

	output := fmt.Sprintf("Successfully deployed %d container(s)", len(reqs))
	return prettycli.PlainOutput{output}, nil
}

func serviceForRequest(req cluster.ContainerRequest) (Service, error) {
	var cc dockerclient.ContainerConfig

	if err := json.Unmarshal(req.CreateOptions, &cc); err != nil {
		return Service{}, err
	}

	return Service{
		Name:            req.Name,
		ContainerConfig: cc,
	}, nil
}

func collectRequests(endpoint cluster.Endpoint, args []string) []cluster.ContainerRequest {
	// TODO: handle error
	go DefaultProxy.Serve(endpoint)
	// TODO: handle error
	defer DefaultProxy.Stop()

	// TODO: handle error
	// TODO: args not passed!
	DefaultComposer.Run(args)
	return DefaultProxy.DrainRequests()
}

func startServices(services []Service, manifests DeploymentManifests, endpoint cluster.Endpoint) error {
	manifestsBlob, err := json.Marshal(manifests)
	if err != nil {
		return err
	}

	for _, svc := range services {
		if svc.ContainerConfig.Labels == nil {
			svc.ContainerConfig.Labels = make(map[string]string)
		}
		svc.ContainerConfig.Labels["zodiacManifest"] = string(manifestsBlob)

		endpoint.StartContainer(svc.Name, svc.ContainerConfig)
	}

	return nil
}
