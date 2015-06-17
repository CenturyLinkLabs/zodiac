package actions

import (
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

var (
	endpointFactory EndpointFactory
)

func init() {
	endpointFactory = func(dockerHost string) (Endpoint, error) {
		c, err := dockerclient.NewDockerClient(dockerHost, nil)
		if err != nil {
			return nil, err
		}

		return &DockerEndpoint{url: dockerHost, client: c}, nil
	}
}

type EndpointFactory func(string) (Endpoint, error)

type Endpoint interface {
	Version() (string, error)
	Name() string
	Host() string
	ResolveImage(string) (string, error)
	StartContainer(name string, cc dockerclient.ContainerConfig) error
	InspectContainer(name string) (*dockerclient.ContainerInfo, error)
	RemoveContainer(name string) error
}

type DockerEndpoint struct {
	url    string
	client dockerclient.Client
}

func (e *DockerEndpoint) Host() string {
	url, _ := url.Parse(e.url)
	return url.Host
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

func (e *DockerEndpoint) StartContainer(name string, cc dockerclient.ContainerConfig) error {
	id, err := e.client.CreateContainer(&cc, name)
	if err != nil {
		log.Fatalf("Problem creating container: ", err)
	}

	log.Infof("%s created as %s", name, id)

	if err := e.client.StartContainer(id, nil); err != nil {
		log.Fatal("problem starting: ", err)
	}
	return nil
}

func (e *DockerEndpoint) ResolveImage(name string) (string, error) {
	imageInfo, err := e.client.InspectImage(name)
	if err != nil {

		if err == dockerclient.ErrNotFound {
			if err := e.client.PullImage(name, nil); err != nil {
				return "", err
			}
			imageInfo, err = e.client.InspectImage(name)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return imageInfo.Id, nil
}

func (e *DockerEndpoint) InspectContainer(name string) (*dockerclient.ContainerInfo, error) {
	return e.client.InspectContainer(name)
}

func (e *DockerEndpoint) RemoveContainer(name string) error {
	//TODO: be more graceful
	return e.client.RemoveContainer(name, true, false)
}
