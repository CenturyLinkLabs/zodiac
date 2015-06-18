package actions

import (
	"errors"
	"testing"

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
	endpointFactory = func(string) (Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.NoError(t, err)
	assert.Equal(t, "Successfully verified endpoint: http://foo.bar", o.ToPrettyOutput())
}

func TestVerify_ErroredOldVersion(t *testing.T) {
	e := mockVerifyEndpoint{version: "1.5.0"}
	endpointFactory = func(string) (Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.EqualError(t, err, "Docker API must be 1.6.0 or above, but it is 1.5.0")
	assert.Nil(t, o)
}

func TestVerify_ErroredCrazyVersion(t *testing.T) {
	e := mockVerifyEndpoint{version: "eleventy-billion"}
	endpointFactory = func(string) (Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.EqualError(t, err, "can't understand Docker version 'eleventy-billion'")
	assert.Nil(t, o)
}

func TestVerify_ErroredAPIError(t *testing.T) {
	e := mockVerifyEndpoint{ErrorForVersion: errors.New("test error")}
	endpointFactory = func(string) (Endpoint, error) {
		return e, nil
	}
	o, err := Verify(Options{})

	assert.EqualError(t, err, "test error")
	assert.Nil(t, o)
}
