package actions

import (
	"encoding/json"
	_ "fmt"
	"testing"

	"github.com/CenturyLinkLabs/zodiac/cluster"
	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
)

type mockRollbackEndpoint struct {
	mockEndpoint
	inspectCallback func(string) (*dockerclient.ContainerInfo, error)
	startCallback   func(string, dockerclient.ContainerConfig) error
}

func (e mockRollbackEndpoint) InspectContainer(nm string) (*dockerclient.ContainerInfo, error) {
	return e.inspectCallback(nm)
}

func (e mockRollbackEndpoint) StartContainer(nm string, cfg dockerclient.ContainerConfig) error {
	return e.startCallback(nm, cfg)
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
