package actions

import (
	"fmt"

	"github.com/CenturyLinkLabs/prettycli"
)

func Teardown(options Options) (prettycli.Output, error) {

	endpoint, err := endpointFactory(options.EndpointOptions)
	if err != nil {
		return nil, err
	}

	reqs, err := collectRequests(options)
	if err != nil {
		return nil, err
	}

	for _, req := range reqs {
		endpoint.RemoveContainer(req.Name)
	}

	output := fmt.Sprintf("Successfully removed %d services and all deployment history", len(reqs))
	return prettycli.PlainOutput{output}, nil
}
