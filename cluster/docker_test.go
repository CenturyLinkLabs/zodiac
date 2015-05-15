package cluster

import (
	"errors"
	"testing"

	"github.com/samalba/dockerclient"
	"github.com/samalba/dockerclient/mockclient"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerEndpointSuccessful(t *testing.T) {
	e, err := NewDockerEndpoint("tcp://example.com:12345")
	assert.NoError(t, err)
	assert.Equal(t, "tcp://example.com:12345", e.Name())
}

func TestNewDockerEndpoint_ErroredBadFormat(t *testing.T) {
	e, err := NewDockerEndpoint("%¡☃🍔!!")
	assert.Nil(t, e)
	assert.EqualError(t, err, `parse %¡☃🍔!!: invalid URL escape "%¡"`)
}

func TestDockerEndpointName(t *testing.T) {
	e, _ := NewDockerEndpoint("tcp://example.com")
	assert.Equal(t, "tcp://example.com", e.Name())
}

func TestDockerEndpointVersion_Successful(t *testing.T) {
	c := mockclient.NewMockClient()
	c.On("Version").Return(&dockerclient.Version{Version: "1.6.1"}, nil).Once()
	e := DockerEndpoint{client: c}

	v, err := e.Version()
	assert.NoError(t, err)
	assert.Equal(t, "1.6.1", v)
	c.AssertExpectations(t)
}

func TestDockerEndpointVersion_ErroredClient(t *testing.T) {
	c := mockclient.NewMockClient()
	c.On("Version").Return(&dockerclient.Version{Version: ""}, errors.New("test error"))
	e := DockerEndpoint{client: c}

	v, err := e.Version()
	assert.Empty(t, "", v)
	assert.EqualError(t, err, "test error")
}
