package cluster

// import "github.com/CenturyLinkLabs/zodiac/proxy"

type Cluster interface {
	Endpoints() []Endpoint
}

type Endpoint interface {
	Version() (string, error)
	Name() string
	// StartContainers([]proxy.ContainerRequest) error
}

type HardcodedCluster []Endpoint

func (c HardcodedCluster) Endpoints() []Endpoint {
	return c
}
