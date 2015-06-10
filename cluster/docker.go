package cluster

import (
	"github.com/CenturyLinkLabs/zodiac/proxy"
	"github.com/samalba/dockerclient"
)

type DockerEndpoint struct {
	url    string
	client dockerclient.Client
}

func NewDockerEndpoint(dockerHost string) (*DockerEndpoint, error) {
	c, err := dockerclient.NewDockerClient(dockerHost, nil)
	if err != nil {
		return nil, err
	}

	return &DockerEndpoint{url: dockerHost, client: c}, nil
}

func (e *DockerEndpoint) Version() (string, error) {
	v, err := e.client.Version()
	if err != nil {
		return "", err
	}

	return v.Version, nil
}

func (e *DockerEndpoint) Name() string {
	return e.url
}

func (e *DockerEndpoint) StartContainers(requests []proxy.ContainerRequest) error {
	// TODO: Implement
	return nil
}
