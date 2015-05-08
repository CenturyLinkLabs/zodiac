package actions

import (
	"encoding/json"
	"fmt"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/discovery"
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

func Rollback(c discovery.Cluster) (prettycli.Output, error) {
	manifests, err := manifestsFromCluster(c)
	if err != nil {
		return prettycli.PlainOutput{}, fmt.Errorf("unable to retrieve manifests from cluster: %s", err)
	}
	if len(manifests) < 2 {
		return prettycli.PlainOutput{"there are no previous deployments to rollback"}, nil
	}

	manifest := manifests[len(manifests)-2]
	log.Infof("found deployment from '%s'", manifest.Time)
	if err := removeContainersForCluster(c); err != nil {
		log.Fatal("unable to remove containers: ", err)
	}

	for _, srv := range manifest.Services {
		srv.ContainerConfig.Image = srv.SHA
	}
	trimmedManifests := manifests[0 : len(manifests)-1]
	trimmedManifestsJSON, err := json.Marshal(trimmedManifests)
	if err != nil {
		log.Fatal("there was a problem building the deployment manifest: ", err)
	}

	for _, ep := range c.Endpoints() {
		client, err := dockerclient.NewDockerClient(ep.URL, nil)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("creating containers on endpoints '%s'", ep.URL)
		for nm, srv := range manifest.Services {
			// TODO Why? Change the way the Config is instantiated?
			if srv.ContainerConfig.Labels == nil {
				srv.ContainerConfig.Labels = make(map[string]string)
			}
			srv.ContainerConfig.Labels["zodiacManifest"] = string(trimmedManifestsJSON)
			id, err := client.CreateContainer(&srv.ContainerConfig, nm)
			if err != nil {
				log.Fatal("problem creating: ", err)
			}
			log.Infof("%s created as %s", nm, id)
			if err := client.StartContainer(id, &dockerclient.HostConfig{}); err != nil {
				log.Fatal("problem starting: ", err)
			}
			log.Infof("started container '%s'", nm)
		}
	}
	return prettycli.PlainOutput{}, nil
}
