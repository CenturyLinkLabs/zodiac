package discovery

import (
	"testing"

	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
)

func TestDockerEndpointClient_Successful(t *testing.T) {
	e := DockerEndpoint{URL: "tcp://example.com:12345"}
	c, ok := e.Client().(*dockerclient.DockerClient)
	if assert.True(t, ok) {
		assert.Equal(t, "http", c.URL.Scheme)
		assert.Equal(t, "example.com:12345", c.URL.Host)
	}
}

// Sadly this tests cannot work because log.Fatal does an os.Exit and there's
// no way to test that it happened, or to abort it. I could do an integration
// test for this, but that feels like cheating. I will defer this and probably
// not come back to it. The cheeseburger stays.
//func TestDockerEndpointClient_ErroredBadFormat(t *testing.T) {
//  e := DockerEndpoint{URL: "%¡☃🍔!!"}
//  e.Client()
//}

func TestDockerEndpointName(t *testing.T) {
	e := DockerEndpoint{URL: "tcp://example.com"}
	assert.Equal(t, "tcp://example.com", e.Name())
}
