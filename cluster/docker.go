package cluster

import (
	"net/url"

	log "github.com/Sirupsen/logrus"
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

	if err := e.client.StartContainer(id, &dockerclient.HostConfig{}); err != nil {
		log.Fatal("problem starting: ", err)
	}
	return nil
}

func (e *DockerEndpoint) ResolveImage(name string) (string, error) {
	imageInfo, err := e.client.InspectImage(name)
	if err != nil {
		if err == dockerclient.ErrNotFound {
			// TODO: authenticaion?
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

//func (e *DockerEndpoint) StartContainers(requests []ContainerRequest) error {
//for _, req := range requests {
//client, err := dockerclient.NewDockerClient(e.Name(), nil)

//id, err := client.CreateContainer(&cc, req.Name)
//if err != nil {
//log.Fatalf("Problem creating container: ", err)
//}

//log.Infof("%s created as %s", req.Name, id)

//if err := client.StartContainer(id, &dockerclient.HostConfig{}); err != nil {
//log.Fatal("problem starting: ", err)
//}
//}
//return nil
//}
