package actions

import (
	"fmt"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/discovery"
	log "github.com/Sirupsen/logrus"
	"github.com/blang/semver"
	"github.com/samalba/dockerclient"
)

var RequredAPIVersion = semver.MustParse("1.6.0")

func Verify(c discovery.Cluster) (prettycli.Output, error) {
	for _, e := range c.Endpoints() {
		if err := verifyEndpoint(e); err != nil {
			return prettycli.PlainOutput{}, err
		}

		log.Infof("endpoint %s verified", e.URL)
	}

	s := fmt.Sprintf("Successfully verified %d endpoint(s)!", len(c.Endpoints()))
	return prettycli.PlainOutput{s}, nil
}

func verifyEndpoint(e discovery.Endpoint) error {
	log.Infof("validating endpoint %s", e.URL)

	client, err := dockerclient.NewDockerClient(e.URL, nil)
	if err != nil {
		return err
	}

	version, err := client.Version()
	if err != nil {
		return err
	}

	semver, err := semver.Make(version.Version)
	if err != nil {
		return fmt.Errorf("Can't understand Docker version '%s'", version.Version)
	}

	if semver.LT(RequredAPIVersion) {
		return fmt.Errorf("Docker API must be %s or above, but it is %s", RequredAPIVersion, semver)
	}

	return nil
}
