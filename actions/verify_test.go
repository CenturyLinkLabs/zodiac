package actions

import (
	"errors"
	"testing"

	"github.com/CenturyLinkLabs/zodiac/discovery"
	"github.com/samalba/dockerclient"
	"github.com/samalba/dockerclient/mockclient"
	"github.com/stretchr/testify/assert"
)

type mockEndpoint struct {
	name   string
	client dockerclient.Client
}

func (e mockEndpoint) Client() dockerclient.Client {
	return e.client
}

func (e mockEndpoint) Name() string {
	return e.name
}

var (
	firstClient  *mockclient.MockClient
	secondClient *mockclient.MockClient
	c            discovery.HardcodedCluster
)

func setup() {
	firstClient = mockclient.NewMockClient()
	secondClient = mockclient.NewMockClient()
	c = discovery.HardcodedCluster{
		mockEndpoint{name: "First", client: firstClient},
		mockEndpoint{name: "Second", client: secondClient},
	}
}

func TestVerify_Success(t *testing.T) {
	setup()
	firstClient.On("Version").Return(&dockerclient.Version{Version: "1.6.1"}, nil)
	secondClient.On("Version").Return(&dockerclient.Version{Version: "1.6.0"}, nil)
	o, err := Verify(c)

	assert.NoError(t, err)
	assert.Equal(t, "Successfully verified 2 endpoint(s)!", o.ToPrettyOutput())
}

func TestVerify_ErroredOldVersion(t *testing.T) {
	setup()
	firstClient.On("Version").Return(&dockerclient.Version{Version: "1.6.1"}, nil)
	secondClient.On("Version").Return(&dockerclient.Version{Version: "1.5.0"}, nil)
	o, err := Verify(c)

	assert.EqualError(t, err, "Docker API must be 1.6.0 or above, but it is 1.5.0")
	assert.Empty(t, o.ToPrettyOutput())
}

func TestVerify_ErroredCrazyVersion(t *testing.T) {
	setup()
	firstClient.On("Version").Return(&dockerclient.Version{Version: "1.6.1"}, nil)
	secondClient.On("Version").Return(&dockerclient.Version{Version: "eleventy-billion"}, nil)
	o, err := Verify(c)

	assert.EqualError(t, err, "can't understand Docker version 'eleventy-billion'")
	assert.Empty(t, o.ToPrettyOutput())
}

func TestVerify_ErroredAPIError(t *testing.T) {
	setup()
	firstClient.On("Version").Return(&dockerclient.Version{Version: "1.6.1"}, nil)
	secondClient.On("Version").Return(&dockerclient.Version{}, errors.New("test error"))
	o, err := Verify(c)

	assert.EqualError(t, err, "test error")
	assert.Empty(t, o.ToPrettyOutput())
}
