package actions

import (
	"encoding/json"
	_ "fmt"
	"testing"

	"github.com/CenturyLinkLabs/zodiac/cluster"
	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
)

type mockDeployEndpoint struct {
	mockEndpoint
	startCallback        func(string, dockerclient.ContainerConfig) error
	resolveImageCallback func(string) (string, error)
}

type mockRollbackEndpoint struct {
	mockEndpoint
	inspectCallback func(string) (*dockerclient.ContainerInfo, error)
	startCallback   func(string, dockerclient.ContainerConfig) error
}

func (e mockDeployEndpoint) StartContainer(nm string, cfg dockerclient.ContainerConfig) error {
	return e.startCallback(nm, cfg)
}

func (e mockDeployEndpoint) ResolveImage(imgNm string) (string, error) {
	return e.resolveImageCallback(imgNm)
}

func (e mockRollbackEndpoint) InspectContainer(nm string) (*dockerclient.ContainerInfo, error) {
	return e.inspectCallback(nm)
}

func (e mockRollbackEndpoint) StartContainer(nm string, cfg dockerclient.ContainerConfig) error {
	return e.startCallback(nm, cfg)
}

type capturedStartParams struct {
	Name   string
	Config dockerclient.ContainerConfig
}

func TestDeploy_Success(t *testing.T) {

	var startCalls []capturedStartParams
	var resolveArgs []string

	DefaultProxy = &mockProxy{
		requests: []cluster.ContainerRequest{
			{
				Name:          "zodiac_foo_1",
				CreateOptions: []byte(`{"Image": "zodiac"}`),
			},
		},
	}
	DefaultComposer = &mockComposer{}

	c := cluster.HardcodedCluster{
		mockDeployEndpoint{
			startCallback: func(nm string, cfg dockerclient.ContainerConfig) error {
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
		},
	}

	o, err := Deploy(c, nil)

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

func TestRollback_Success(t *testing.T) {

	var startCalls []capturedStartParams
	previousManis := []DeploymentManifest{
		{
			Services: []Service{
				{
					Name: "oldService",
					ContainerConfig: dockerclient.ContainerConfig{
						Image: "oldimage",
					},
				},
			},
		},
		{
			Services: []Service{
				{
					Name:            "newService",
					ContainerConfig: dockerclient.ContainerConfig{},
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
		requests: []cluster.ContainerRequest{
			{
				Name:          "zodiac_foo_1",
				CreateOptions: []byte(`{"Image": "zodiac"}`),
			},
		},
	}
	DefaultComposer = &mockComposer{}

	c := cluster.HardcodedCluster{
		mockRollbackEndpoint{
			inspectCallback: func(nm string) (*dockerclient.ContainerInfo, error) {
				return &ci, nil
			},
			startCallback: func(nm string, cfg dockerclient.ContainerConfig) error {
				startCalls = append(startCalls, capturedStartParams{
					Name:   nm,
					Config: cfg,
				})
				return nil
			},
		},
	}

	o, err := Rollback(c, nil)

	assert.NoError(t, err)
	assert.Len(t, startCalls, 1)
	mostRecentCall := startCalls[0]
	assert.Equal(t, "oldService", mostRecentCall.Name)
	assert.Equal(t, "oldimage", mostRecentCall.Config.Image)
	assert.NotEmpty(t, mostRecentCall.Config.Labels["zodiacManifest"])
	assert.Equal(t, "Successfully rolled back to 1 container(s)", o.ToPrettyOutput())

	dms := DeploymentManifests{}
	err = json.Unmarshal([]byte(mostRecentCall.Config.Labels["zodiacManifest"]), &dms)
	assert.NoError(t, err)
	assert.Len(t, dms, 3)
	dm := dms[2]
	assert.NotEqual(t, "", dm.DeployedAt)
}
