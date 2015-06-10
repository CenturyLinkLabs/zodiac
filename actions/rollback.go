package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/cluster"
)

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
