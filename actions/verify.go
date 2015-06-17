package actions

import (
	"fmt"

	"github.com/CenturyLinkLabs/prettycli"
	log "github.com/Sirupsen/logrus"
	"github.com/blang/semver"
)

var RequredAPIVersion = semver.MustParse("1.6.0")

func Verify(options Options) (prettycli.Output, error) {

	endpoint, err := endpointFactory(options.Flags["endpoint"])
	if err != nil {
		return nil, err
	}

	log.Infof("validating endpoint %s", endpoint.Name())

	if err := verifyEndpoint(endpoint); err != nil {
		return prettycli.PlainOutput{}, err
	}

	s := fmt.Sprintf("Successfully verified endpoint: %s", endpoint.Name())
	return prettycli.PlainOutput{s}, nil
}

func verifyEndpoint(e Endpoint) error {
	version, err := e.Version()
	if err != nil {
		return err
	}

	log.Infof("%s reported version %s", e.Name(), version)
	semver, err := semver.Make(version)
	if err != nil {
		return fmt.Errorf("can't understand Docker version '%s'", version)
	}

	if semver.LT(RequredAPIVersion) {
		return fmt.Errorf("Docker API must be %s or above, but it is %s", RequredAPIVersion, semver)
	}

	return nil
}
