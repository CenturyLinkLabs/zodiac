package actions

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/CenturyLinkLabs/prettycli"
)

func Rollback(options Options) (prettycli.Output, error) {

	endpoint, err := endpointFactory(options.Flags["endpoint"])
	if err != nil {
		return nil, err
	}

	reqs, err := collectRequests(options)
	if err != nil {
		return nil, err
	}

	// Get most recent deployment's manifests
	manifests, err := getDeploymentManifests(reqs, endpoint)
	if err != nil {
		if len(manifests) <= 1 {
			return nil, errors.New("There are no previous deployments")
		}
		return nil, err
	}

	if len(manifests) <= 1 {
		return nil, errors.New("There are no previous deployments")
	}

	newDeployment, deploymentID, err := fetchTarget(manifests, options.Args)
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
	manifests[len(manifests)-1].DeployedAt = time.Now().Format(BasicDateTime)

	if err := startServices(newDeployment.Services, manifests, endpoint); err != nil {
		return nil, err
	}

	output := fmt.Sprintf("Successfully rolled back to deployment: %d", deploymentID)
	return prettycli.PlainOutput{output}, nil
}

func fetchTarget(manifests DeploymentManifests, args []string) (DeploymentManifest, int, error) {
	var target int
	if len(args) == 0 {
		target = len(manifests) - 2
	} else {
		var err error
		target, err = strconv.Atoi(args[0])
		if err != nil {
			return DeploymentManifest{}, -1, err
		}
		target--
	}

	if (target < 0) || (target >= len(manifests)) {
		return DeploymentManifest{}, -1, errors.New("The specified index does not exist")
	}

	return manifests[target], target + 1, nil
}
