package actions

import (
	"github.com/CenturyLinkLabs/zodiac/proxy"
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
	requests []proxy.ContainerRequest
}

func (p mockProxy) Serve() error {
	return nil
}

func (p mockProxy) Stop() error {
	return nil
}

func (p mockProxy) GetRequests() []proxy.ContainerRequest {
	return p.requests
}

type mockComposer struct{}

func (c *mockComposer) Run(flags map[string]string) error {
	return nil
}

type capturedStartParams struct {
	Name   string
	Config dockerclient.ContainerConfig
}
