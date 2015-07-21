package actions

import (
	"fmt"
	"strings"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/endpoint"
	log "github.com/Sirupsen/logrus"
	"github.com/blang/semver"
)

var RequiredDockerAPIVersion = semver.MustParse("1.6.0")
var RequiredSwarmAPIVersion = semver.MustParse("0.3.0")

func Verify(options Options) (prettycli.Output, error) {

	endpoint, err := endpointFactory(options.EndpointOptions)
	if err != nil {
		return nil, err
	}

	log.Infof("Validating endpoint %s", endpoint.Name())

	if err := verifyEndpoint(endpoint); err != nil {
		return nil, err
	}

	s := fmt.Sprintf("Successfully verified endpoint: %s", endpoint.Name())
	return prettycli.PlainOutput{s}, nil
}

func verifyEndpoint(e endpoint.Endpoint) error {
	version, err := e.Version()
	if err != nil {
		return err
	}

	log.Infof("%s reported version %s", e.Name(), version)

	isSwarm := false
	if strings.HasPrefix(version, "swarm/") {
		isSwarm = true
		parts := strings.Split(version, "/")
		version = parts[1]
	}

	semver, err := semver.Make(version)
	if err != nil {
		return fmt.Errorf("can't understand version '%s'", version)
	}

	if isSwarm && semver.LT(RequiredSwarmAPIVersion) {
		return fmt.Errorf("Swarm API must be %s or above, but it is %s", RequiredSwarmAPIVersion, semver)
	}

	if !isSwarm && semver.LT(RequiredDockerAPIVersion) {
		return fmt.Errorf("Docker API must be %s or above, but it is %s", RequiredDockerAPIVersion, semver)
	}

	return nil
}
