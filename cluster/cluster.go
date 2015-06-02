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
	StartContainers([]ContainerRequest) error
}

type HardcodedCluster []Endpoint

func (c HardcodedCluster) Endpoints() []Endpoint {
	return c
}
