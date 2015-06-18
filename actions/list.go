package actions

import (
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

	reqs, err := collectRequests(options)
	if err != nil {
		return nil, err
	}

	// Get most recent deployment's manifests
	manifests, err := getDeploymentManifests(reqs, endpoint)
	if err != nil {
		return nil, err
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
