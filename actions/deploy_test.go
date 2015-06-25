package actions

import (
	"encoding/json"
	_ "fmt"
	"testing"

	"github.com/CenturyLinkLabs/zodiac/proxy"
	"github.com/stretchr/testify/assert"
)

type mockDeployEndpoint struct {
	mockEndpoint
	startCallback        func(string, ContainerConfig) error
	resolveImageCallback func(string) (string, error)
}

func (e mockDeployEndpoint) StartContainer(nm string, cfg ContainerConfig) error {
	return e.startCallback(nm, cfg)
}

func (e mockDeployEndpoint) ResolveImage(imgNm string) (string, error) {
	return e.resolveImageCallback(imgNm)
}

func TestDeploy_Success(t *testing.T) {

	var startCalls []capturedStartParams
	var resolveArgs []string

	DefaultProxy = &mockProxy{
		requests: []proxy.ContainerRequest{
			{
				Name:          "zodiac_foo_1",
				CreateOptions: []byte(`{"Image": "zodiac"}`),
			},
		},
	}
	DefaultComposer = &mockComposer{}

	e := mockDeployEndpoint{
		startCallback: func(nm string, cfg ContainerConfig) error {
			startCalls = append(startCalls, capturedStartParams{
				Name:   nm,
				Config: cfg,
			})
			return nil
		},
		resolveImageCallback: func(imgNm string) (string, error) {
			resolveArgs = append(resolveArgs, imgNm)
			return "xyz321", nil
		},
	}

	endpointFactory = func(string) (Endpoint, error) {
		return e, nil
	}

	o, err := Deploy(Options{})

	assert.NoError(t, err)
	assert.Len(t, startCalls, 1)
	mostRecentCall := startCalls[0]
	assert.Equal(t, "zodiac_foo_1", mostRecentCall.Name)
	assert.Equal(t, "xyz321", mostRecentCall.Config.Image)
	assert.Equal(t, []string{"zodiac"}, resolveArgs)
	assert.NotEmpty(t, mostRecentCall.Config.Labels["zodiacManifest"])
	assert.Equal(t, "Successfully deployed 1 container(s)", o.ToPrettyOutput())

	dms := DeploymentManifests{}
	err = json.Unmarshal([]byte(mostRecentCall.Config.Labels["zodiacManifest"]), &dms)
	assert.NoError(t, err)
	assert.Len(t, dms, 1)
	dm := dms[0]
	assert.NotEqual(t, "", dm.DeployedAt)
	assert.Equal(t, "xyz321", dm.Services[0].ContainerConfig.Image)
}
