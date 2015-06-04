package cluster

import (
	"github.com/samalba/dockerclient"
)

type ContainerRequest struct {
	Name          string
	CreateOptions []byte
	Config        dockerclient.ContainerConfig
}

type Cluster interface {
	Endpoints() []Endpoint
}

type Endpoint interface {
	Version() (string, error)
	Name() string
	Host() string
	ResolveImage(string) (string, error)
	//StartContainers([]ContainerRequest) error
	StartContainer(name string, cc dockerclient.ContainerConfig) error
}

type HardcodedCluster []Endpoint

func (c HardcodedCluster) Endpoints() []Endpoint {
	return c
}
