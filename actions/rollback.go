package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/cluster"
)

func Rollback(c cluster.Cluster, options Options) (prettycli.Output, error) {
	var reqs []cluster.ContainerRequest

	for _, endpoint := range c.Endpoints() {

		reqs = collectRequests(endpoint, options)
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

		if len(manifests) <= 1 {
			return nil, errors.New("There are no previous deployments")
		}

		newDeployment, err := fetchTarget(manifests, options.Args)
		if err != nil {
			return nil, err
		}

		// shut down current deployment
		currentDeployment := manifests[len(manifests)-1]

		for _, svc := range currentDeployment.Services {
			endpoint.RemoveContainer(svc.Name)
		}

		manifests = append(manifests, newDeployment)
		newDeployment = manifests[len(manifests)-1]
		manifests[len(manifests)-1].DeployedAt = time.Now().Format(time.RFC3339)

		startServices(newDeployment.Services, manifests, endpoint)
	}

	output := fmt.Sprintf("Successfully rolled back to %d container(s)", len(reqs))
	return prettycli.PlainOutput{output}, nil
}

func fetchTarget(manifests DeploymentManifests, args []string) (DeploymentManifest, error) {
	var target int
	if len(args) == 0 {
		target = len(manifests) - 2
	} else {
		var err error
		target, err = strconv.Atoi(args[0])
		if err != nil {
			return DeploymentManifest{}, err
		}
		target--
	}

	if (target < 0) || (target >= len(manifests)) {
		return DeploymentManifest{}, errors.New("The specified index does not exist")
	}

	return manifests[target], nil
}
