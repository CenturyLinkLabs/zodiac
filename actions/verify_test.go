package actions

import (
	"errors"
	"testing"

	"github.com/CenturyLinkLabs/zodiac/cluster"
	"github.com/stretchr/testify/assert"
)

type mockVerifyEndpoint struct {
	mockEndpoint
	ErrorForVersion error
	version         string
}

func (e mockVerifyEndpoint) Version() (string, error) {
	if e.ErrorForVersion != nil {
		return "", e.ErrorForVersion
	}
	return e.version, nil
}

func TestVerify_Success(t *testing.T) {
	c := cluster.HardcodedCluster{
		mockVerifyEndpoint{version: "1.6.1"},
		mockVerifyEndpoint{version: "1.6.0"},
	}
	o, err := Verify(c, Options{})

	assert.NoError(t, err)
	assert.Equal(t, "Successfully verified 2 endpoint(s)!", o.ToPrettyOutput())
}

func TestVerify_ErroredOldVersion(t *testing.T) {
	c := cluster.HardcodedCluster{
		mockVerifyEndpoint{version: "1.6.1"},
		mockVerifyEndpoint{version: "1.5.0"},
	}
	o, err := Verify(c, Options{})

	assert.EqualError(t, err, "Docker API must be 1.6.0 or above, but it is 1.5.0")
	assert.Empty(t, o.ToPrettyOutput())
}

func TestVerify_ErroredCrazyVersion(t *testing.T) {
	c := cluster.HardcodedCluster{
		mockVerifyEndpoint{version: "1.6.1"},
		mockVerifyEndpoint{version: "eleventy-billion"},
	}
	o, err := Verify(c, Options{})

	assert.EqualError(t, err, "can't understand Docker version 'eleventy-billion'")
	assert.Empty(t, o.ToPrettyOutput())
}

func TestVerify_ErroredAPIError(t *testing.T) {
	c := cluster.HardcodedCluster{
		mockVerifyEndpoint{version: "1.6.1"},
		mockVerifyEndpoint{ErrorForVersion: errors.New("test error")},
	}
	o, err := Verify(c, Options{})

	assert.EqualError(t, err, "test error")
	assert.Empty(t, o.ToPrettyOutput())
}
