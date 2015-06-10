package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/cluster"
)

func List(c cluster.Cluster, args []string) (prettycli.Output, error) {
	var reqs []cluster.ContainerRequest

	output := prettycli.ListOutput{
		Labels: []string{"ID", "Deploy Date", "Services"},
	}

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

		// Iterate backwards from most recent mani to oldest
		for i := len(manifests) - 1; i >= 0; i-- {
			mani := manifests[i]
			var serviceList []string
			for _, svc := range mani.Services {
				serviceList = append(serviceList, svc.Name)
			}

			output.AddRow(map[string]string{
				"ID":          strconv.Itoa(i + 1),
				"Deploy Date": mani.DeployedAt,
				"Services":    strings.Join(serviceList, ", "),
			})
		}

	}

	return output, nil
}
