package actions

import (
	"github.com/CenturyLinkLabs/zodiac/cluster"
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

func init() {
	log.SetLevel(log.FatalLevel)
}

type mockEndpoint struct{}

func (e mockEndpoint) Host() string {
	return "fakeHost"
}

func (e mockEndpoint) Name() string {
	return "fakeName that is really a URI"
}

func (e mockEndpoint) StartContainer(string, dockerclient.ContainerConfig) error {
	return nil
}

func (e mockEndpoint) ResolveImage(imgNm string) (string, error) {
	return "abc123", nil
}

func (e mockEndpoint) InspectContainer(name string) (*dockerclient.ContainerInfo, error) {
	return &dockerclient.ContainerInfo{}, nil
}

func (e mockEndpoint) RemoveContainer(name string) error {
	return nil
}

func (e mockEndpoint) Version() (string, error) {
	return "1.0", nil
}

type mockProxy struct {
	requests []cluster.ContainerRequest
}

func (p mockProxy) Serve(ep cluster.Endpoint) error {
	return nil
}

func (p mockProxy) Stop() error {
	return nil
}

func (p mockProxy) DrainRequests() []cluster.ContainerRequest {
	return p.requests
}

type mockComposer struct{}

func (c *mockComposer) Run(args []string) error {
	return nil
}
