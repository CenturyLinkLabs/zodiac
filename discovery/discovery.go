package discovery

import (
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

type Cluster interface {
	Endpoints() []Endpoint
}

type Endpoint interface {
	Client() dockerclient.Client
	Name() string
}

type HardcodedCluster []Endpoint

func (c HardcodedCluster) Endpoints() []Endpoint {
	return c
}

type DockerEndpoint struct {
	URL    string
	client *dockerclient.DockerClient
}

func (e *DockerEndpoint) Client() dockerclient.Client {
	if e.client == nil {
		c, err := dockerclient.NewDockerClient(e.URL, nil)
		if err != nil {
			log.Fatal("there was a problem building the Docker client: ", err)
		}
		e.client = c
	}

	return e.client
}

func (e *DockerEndpoint) Name() string {
	return e.URL
}
