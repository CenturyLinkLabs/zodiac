package actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CenturyLinkLabs/prettycli"
)

func List(options Options) (prettycli.Output, error) {

	endpoint, err := endpointFactory(options.EndpointOptions)
	if err != nil {
		return nil, err
	}

	output := prettycli.ListOutput{
		Labels: []string{"Active", "ID", "Deploy Date", "Services", "Message"},
	}

	reqs, err := collectRequests(options, true)
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

		var isActive string
		if i == (len(manifests) - 1) {
			isActive = "*"
		}

		output.AddRow(map[string]string{
			"Active":      isActive,
			"ID":          strconv.Itoa(i + 1),
			"Deploy Date": mani.DeployedAt,
			"Services":    strings.Join(serviceList, ", "),
			"Message":     truncate(mani.Message, 72),
		})
	}

	return output, nil
}

func truncate(msg string, length int) string {

	if len(msg) <= length {
		return msg
	}

	return fmt.Sprintf("%s...", msg[0:length-1])
}
