package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/cluster"
	"github.com/samalba/dockerclient"
)

func Deploy(c cluster.Cluster, options Options) (prettycli.Output, error) {

	var reqs []cluster.ContainerRequest

	for _, endpoint := range c.Endpoints() {

		reqs = collectRequests(endpoint, options)

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
