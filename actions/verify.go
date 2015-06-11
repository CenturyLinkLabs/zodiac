package actions

import (
	"fmt"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/cluster"
	log "github.com/Sirupsen/logrus"
	"github.com/blang/semver"
)

var RequredAPIVersion = semver.MustParse("1.6.0")

func Verify(c cluster.Cluster, options Options) (prettycli.Output, error) {
	for _, e := range c.Endpoints() {
		log.Infof("validating endpoint %s", e.Name())

		if err := verifyEndpoint(e); err != nil {
			return prettycli.PlainOutput{}, err
		}
	}

	s := fmt.Sprintf("Successfully verified %d endpoint(s)!", len(c.Endpoints()))
	return prettycli.PlainOutput{s}, nil
}

func verifyEndpoint(e cluster.Endpoint) error {
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
