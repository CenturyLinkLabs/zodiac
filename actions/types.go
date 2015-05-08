package actions

import (
	"encoding/json"

	"github.com/CenturyLinkLabs/zodiac/discovery"
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

type DeploymentManifests []DeploymentManifest
type DeploymentManifest struct {
	Services map[string]string
	Time     string
}

func NewDeploymentManifest() DeploymentManifest {
	return DeploymentManifest{
		Services: make(map[string]string),
	}
}

func manifestsFromCluster(c discovery.Cluster) (DeploymentManifests, error) {
	var oldManifests DeploymentManifests
	for _, ep := range c.Endpoints() {
		log.Infof("checking deployments for '%s'", ep.URL)
		client, err := dockerclient.NewDockerClient(ep.URL, nil)
		if err != nil {
			log.Fatal("failed to create docker client: ", err)
		}
		cts, err := client.ListContainers(true, false, "")
		if err != nil {
			log.Fatal("unable to list containers: ", err)
		}

		for _, ct := range cts {
			if err := json.Unmarshal([]byte(ct.Labels["zodiacManifest"]), &oldManifests); err != nil {
				log.Fatal("unable to unmarshal manifest: ", err)
			}
		}
	}
	return oldManifests, nil
}

func removeContainers(c discovery.Cluster) error {
	for _, ep := range c.Endpoints() {
		log.Infof("checking deployments for '%s'", ep.URL)
		client, err := dockerclient.NewDockerClient(ep.URL, nil)
		if err != nil {
			log.Fatal("failed to create docker client: ", err)
		}
		cts, err := client.ListContainers(true, false, "")
		if err != nil {
			log.Fatal("unable to list containers: ", err)
		}

		for _, ct := range cts {
			log.Infof("removing container '%s'", ct.Id)
			// TODO: should we be removing the associated volumes?
			if err := client.RemoveContainer(ct.Id, true, false); err != nil {
				log.Fatal("unable to destroy container: ", err)
			}
		}
	}
	return nil
}
