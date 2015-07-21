package actions

import (
	"errors"
	"testing"

	"github.com/CenturyLinkLabs/zodiac/endpoint"
	"github.com/stretchr/testify/assert"
)

type mockVerifyEndpoint struct {
	mockEndpoint
	ErrorForVersion error
	version         string
	url             string
}

func (e mockVerifyEndpoint) Version() (string, error) {
	if e.ErrorForVersion != nil {
		return "", e.ErrorForVersion
	}
	return e.version, nil
}

func (e mockVerifyEndpoint) Name() string {
	return e.url
}

func TestVerify_Success(t *testing.T) {
	e := mockVerifyEndpoint{version: "1.6.1", url: "http://foo.bar"}
	endpointFactory = func(endpoint.EndpointOptions) (endpoint.Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.NoError(t, err)
	assert.Equal(t, "Successfully verified endpoint: http://foo.bar", o.ToPrettyOutput())
}

func TestVerify_ErroredOldVersion(t *testing.T) {
	e := mockVerifyEndpoint{version: "1.5.0"}
	endpointFactory = func(endpoint.EndpointOptions) (endpoint.Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.EqualError(t, err, "Docker API must be 1.6.0 or above, but it is 1.5.0")
	assert.Nil(t, o)
}

func TestVerify_ErroredCrazyVersion(t *testing.T) {
	e := mockVerifyEndpoint{version: "eleventy-billion"}
	endpointFactory = func(endpoint.EndpointOptions) (endpoint.Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.EqualError(t, err, "can't understand version 'eleventy-billion'")
	assert.Nil(t, o)
}

func TestVerify_ErroredAPIError(t *testing.T) {
	e := mockVerifyEndpoint{ErrorForVersion: errors.New("test error")}
	endpointFactory = func(endpoint.EndpointOptions) (endpoint.Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.EqualError(t, err, "test error")
	assert.Nil(t, o)
}

func TestVerify_SwarmSuccess(t *testing.T) {
	e := mockVerifyEndpoint{version: "swarm/1.6.1", url: "http://foo.bar"}
	endpointFactory = func(endpoint.EndpointOptions) (endpoint.Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.NoError(t, err)
	assert.Equal(t, "Successfully verified endpoint: http://foo.bar", o.ToPrettyOutput())
}

func TestVerify_SwarmOldVersion(t *testing.T) {
	e := mockVerifyEndpoint{version: "swarm/0.2.1", url: "http://foo.bar"}
	endpointFactory = func(endpoint.EndpointOptions) (endpoint.Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.EqualError(t, err, "Swarm API must be 0.3.0 or above, but it is 0.2.1")
	assert.Nil(t, o)
}
