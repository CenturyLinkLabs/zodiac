package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/CenturyLinkLabs/prettycli"
)

func List(options Options) (prettycli.Output, error) {

	endpoint, err := endpointFactory(options.Flags["endpoint"])
	if err != nil {
		return nil, err
	}

	output := prettycli.ListOutput{
		Labels: []string{"ID", "Deploy Date", "Services", "Message"},
	}

	reqs := collectRequests(options)
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
			"Message":     mani.Message,
		})
	}

	return output, nil
}
