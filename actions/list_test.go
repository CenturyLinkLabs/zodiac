package actions

import (
	"encoding/json"
	_ "fmt"
	"testing"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/proxy"
	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
)

type mockListEndpoint struct {
	mockEndpoint
	inspectCallback func(string) (*dockerclient.ContainerInfo, error)
}

func (e mockListEndpoint) InspectContainer(nm string) (*dockerclient.ContainerInfo, error) {
	return e.inspectCallback(nm)
}

func TestList_Success(t *testing.T) {
	oldServiceConfig := ContainerConfig{}
	oldServiceConfig.Image = "oldimage"

	previousManis := []DeploymentManifest{
		{
			Services: []Service{
				{
					Name:            "oldService",
					ContainerConfig: oldServiceConfig,
				},
			},
		},
		{
			DeployedAt: "yesterday",
			Message:    "initial release for deployment to production, but don't worry you won't see all this text in the list view",
			Services: []Service{
				{
					Name:            "newService",
					ContainerConfig: ContainerConfig{},
				},
				{
					Name:            "Another service",
					ContainerConfig: ContainerConfig{},
				},
			},
		},
	}
	previousManisBlob, _ := json.Marshal(previousManis)

	ci := dockerclient.ContainerInfo{
		Config: &dockerclient.ContainerConfig{
			Labels: map[string]string{
				"zodiacManifest": string(previousManisBlob),
			},
		},
	}

	DefaultProxy = &mockProxy{
		requests: []proxy.ContainerRequest{
			{
				Name:          "zodiac_foo_1",
				CreateOptions: []byte(`{"Image": "zodiac"}`),
			},
		},
	}
	DefaultComposer = &mockComposer{}

	e := mockListEndpoint{
		inspectCallback: func(nm string) (*dockerclient.ContainerInfo, error) {
			return &ci, nil
		},
	}

	endpointFactory = func(EndpointOptions) (Endpoint, error) {
		return e, nil
	}

	o, err := List(Options{})

	output, _ := o.(prettycli.ListOutput)

	assert.NoError(t, err)
	assert.Len(t, output.Labels, 4)
	assert.Len(t, output.Rows, 2)
	assert.Equal(t, "ID", output.Labels[0])
	assert.Equal(t, "Deploy Date", output.Labels[1])
	assert.Equal(t, "Services", output.Labels[2])
	assert.Equal(t, "2", output.Rows[0]["ID"])
	assert.Equal(t, "yesterday", output.Rows[0]["Deploy Date"])
	assert.Equal(t, "newService, Another service", output.Rows[0]["Services"])
	assert.Equal(t, "initial release for deployment to production, but don't worry you won't...", output.Rows[0]["Message"])
}
