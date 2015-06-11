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
	removeCallback  func(string) error
}

func (e mockRollbackEndpoint) InspectContainer(nm string) (*dockerclient.ContainerInfo, error) {
	return e.inspectCallback(nm)
}

func (e mockRollbackEndpoint) StartContainer(nm string, cfg dockerclient.ContainerConfig) error {
	return e.startCallback(nm, cfg)
}

func (e mockRollbackEndpoint) RemoveContainer(nm string) error {
	if e.removeCallback != nil {
		return e.removeCallback(nm)
	} else {
		return nil
	}
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

	o, err := Rollback(c, Options{})

	assert.NoError(t, err)
	assert.Len(t, startCalls, 1)
	mostRecentCall := startCalls[0]
	assert.Equal(t, "oldService", mostRecentCall.Name)
	assert.Equal(t, "oldimage", mostRecentCall.Config.Image)
	assert.NotEmpty(t, mostRecentCall.Config.Labels["zodiacManifest"])
	assert.Equal(t, "Successfully rolled back to deployment: 1", o.ToPrettyOutput())

	dms := DeploymentManifests{}
	err = json.Unmarshal([]byte(mostRecentCall.Config.Labels["zodiacManifest"]), &dms)
	assert.NoError(t, err)
	assert.Len(t, dms, 3)
	dm := dms[2]
	assert.NotEqual(t, "", dm.DeployedAt)
}

func TestRollbackWithID_Success(t *testing.T) {

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

	_, err := Rollback(c, Options{
		Args: []string{"1"},
	})

	assert.NoError(t, err)
	assert.Len(t, startCalls, 1)
	mostRecentCall := startCalls[0]
	assert.Equal(t, "oldService", mostRecentCall.Name)
	assert.Equal(t, "oldimage", mostRecentCall.Config.Image)
}

func TestRollbackWithNoPreviousDeployment_Error(t *testing.T) {

	var startCalls []capturedStartParams
	var removeCalls []string
	previousManis := []DeploymentManifest{
		{
			Services: []Service{
				{
					Name:            "OnlyDeployment",
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
			removeCallback: func(nm string) error {
				removeCalls = append(removeCalls, nm)
				return nil
			},
		},
	}

	_, err := Rollback(c, Options{})

	assert.EqualError(t, err, "There are no previous deployments")
	assert.Len(t, startCalls, 0)
	assert.Len(t, removeCalls, 0)
}

func TestRollbackWithNonexistingID_Error(t *testing.T) {

	var startCalls []capturedStartParams
	var removeCalls []string
	previousManis := []DeploymentManifest{
		{
			Services: []Service{
				{
					Name:            "First Deplyment",
					ContainerConfig: dockerclient.ContainerConfig{},
				},
			},
		},
		{
			Services: []Service{
				{
					Name:            "Second Deployment",
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
			removeCallback: func(nm string) error {
				removeCalls = append(removeCalls, nm)
				return nil
			},
		},
	}

	_, err := Rollback(c, Options{
		Args: []string{"3"},
	})

	assert.EqualError(t, err, "The specified index does not exist")
	assert.Len(t, startCalls, 0)
	assert.Len(t, removeCalls, 0)
}
