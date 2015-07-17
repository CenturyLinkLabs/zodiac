package actions

import (
	_ "fmt"
	"testing"

	"github.com/CenturyLinkLabs/zodiac/endpoint"
	"github.com/CenturyLinkLabs/zodiac/proxy"
	"github.com/stretchr/testify/assert"
)

type mockTeardownEndpoint struct {
	mockEndpoint
	removeCallback func(string) error
}

func (e mockTeardownEndpoint) RemoveContainer(nm string) error {
	if e.removeCallback != nil {
		return e.removeCallback(nm)
	} else {
		return nil
	}
}

func TestTeardown_Success(t *testing.T) {

	var removeCalls []string

	DefaultProxy = &mockProxy{
		requests: []proxy.ContainerRequest{
			{
				Name: "zodiac_foo_1",
			},
			{
				Name: "zodiac_boo_2",
			},
		},
	}
	DefaultComposer = &mockComposer{}

	e := mockTeardownEndpoint{
		removeCallback: func(nm string) error {
			removeCalls = append(removeCalls, nm)
			return nil
		},
	}

	endpointFactory = func(endpoint.EndpointOptions) (endpoint.Endpoint, error) {
		return e, nil
	}

	o, err := Teardown(Options{})

	assert.NoError(t, err)
	assert.Len(t, removeCalls, 2)
	assert.Equal(t, "zodiac_foo_1", removeCalls[0])
	assert.Equal(t, "zodiac_boo_2", removeCalls[1])
	assert.Equal(t, "Successfully removed 2 services and all deployment history", o.ToPrettyOutput())
}
