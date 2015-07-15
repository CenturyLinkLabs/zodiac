package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/proxy"
)

func Deploy(options Options) (prettycli.Output, error) {
	fmt.Println("Deploying your application...")

	endpoint, err := endpointFactory(options.EndpointOptions)
	if err != nil {
		return nil, err
	}

	reqs, err := collectRequests(options, false)
	if err != nil {
		return nil, err
	}

	dm := DeploymentManifest{
		Services:   []Service{},
		DeployedAt: time.Now().Format(BasicDateTime),
		Message:    options.Flags["message"],
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

	if err = startServices(dm.Services, manifests, endpoint); err != nil {
		return nil, err
	}

	output := fmt.Sprintf("Successfully deployed %d container(s)", len(reqs))
	return prettycli.PlainOutput{output}, nil
}

func serviceForRequest(req proxy.ContainerRequest) (Service, error) {
	var cc ContainerConfig

	if err := json.Unmarshal(req.CreateOptions, &cc); err != nil {
		return Service{}, err
	}

	return Service{
		Name:            req.Name,
		ContainerConfig: cc,
	}, nil
}
