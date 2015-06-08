package cluster

import (
	"errors"
	"testing"

	"github.com/samalba/dockerclient"
	"github.com/samalba/dockerclient/mockclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestResolveImage_WhenImageExisting(t *testing.T) {
	c := mockclient.NewMockClient()
	c.On("InspectImage", "Foo").Return(&dockerclient.ImageInfo{Id: "ytu678"}, nil)
	e := DockerEndpoint{client: c}
	imageID, err := e.ResolveImage("Foo")

	assert.NoError(t, err)
	assert.Equal(t, "ytu678", imageID)
}

func TestResolveImage_WhenImageDoesNotExist(t *testing.T) {
	c := mockclient.NewMockClient()
	c.On("InspectImage", "Foo").Return(&dockerclient.ImageInfo{}, dockerclient.ErrNotFound).Once()
	c.On("PullImage", "Foo", mock.Anything).Return(nil)
	c.On("InspectImage", "Foo").Return(&dockerclient.ImageInfo{Id: "yui890"}, nil).Once()
	e := DockerEndpoint{client: c}
	imageID, err := e.ResolveImage("Foo")

	assert.NoError(t, err)
	assert.Equal(t, "yui890", imageID)
	c.AssertExpectations(t)
}

func TestResolveImage_WhenInitialInspectErrors(t *testing.T) {
	c := mockclient.NewMockClient()
	c.On("InspectImage", "Foo").Return(&dockerclient.ImageInfo{}, errors.New("oops"))

	e := DockerEndpoint{client: c}
	imageID, err := e.ResolveImage("Foo")

	assert.Equal(t, "", imageID)
	assert.EqualError(t, err, "oops")
	c.AssertExpectations(t)
}

func TestResolveImage_WhenPullErrors(t *testing.T) {
	c := mockclient.NewMockClient()
	c.On("InspectImage", "Foo").Return(&dockerclient.ImageInfo{}, dockerclient.ErrNotFound).Once()
	c.On("PullImage", "Foo", mock.Anything).Return(errors.New("uh-oh"))

	e := DockerEndpoint{client: c}
	imageID, err := e.ResolveImage("Foo")

	assert.Equal(t, "", imageID)
	assert.EqualError(t, err, "uh-oh")
	c.AssertExpectations(t)
}

func TestResolveImage_WhenSecondInspectErrors(t *testing.T) {
	c := mockclient.NewMockClient()
	c.On("InspectImage", "Foo").Return(&dockerclient.ImageInfo{}, dockerclient.ErrNotFound).Once()
	c.On("PullImage", "Foo", mock.Anything).Return(nil)
	c.On("InspectImage", "Foo").Return(&dockerclient.ImageInfo{}, errors.New("whoops")).Once()

	e := DockerEndpoint{client: c}
	imageID, err := e.ResolveImage("Foo")

	assert.Equal(t, "", imageID)
	assert.EqualError(t, err, "whoops")
	c.AssertExpectations(t)
}
