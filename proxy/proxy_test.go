package proxy

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
)

func TestGetRequests_WithNoErrors(t *testing.T) {
	proxy := HTTPProxy{
		containerRequests: []ContainerRequest{
			{Name: "foo"},
			{Name: "bar"},
		},
	}

	reqs, err := proxy.GetRequests()
	assert.NoError(t, err)
	assert.Len(t, reqs, 2)
	assert.Equal(t, "foo", reqs[0].Name)
	assert.Equal(t, "bar", reqs[1].Name)
}

func TestGetRequests_WithErrors(t *testing.T) {
	proxy := HTTPProxy{
		containerRequests: []ContainerRequest{
			{Name: "foo"},
			{Name: "bar"},
		},
		errors: []error{errors.New("oops")},
	}

	reqs, err := proxy.GetRequests()
	assert.Error(t, err)
	assert.Nil(t, reqs)
}

func TestCreate(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61900",
	}

	go proxy.Serve()
	defer proxy.Stop()

	resp, err := http.Post("http://localhost:61900/v1.15/containers/create?name=foo", "", strings.NewReader("bar"))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"Id":"doesnt_matter", "Warnings":[]}`, string(body))

	assert.Len(t, proxy.containerRequests, 1)
	assert.Equal(t, "foo", proxy.containerRequests[0].Name)
	assert.Equal(t, "bar", string(proxy.containerRequests[0].CreateOptions))
}

func TestInspect(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61901",
	}

	go proxy.Serve()
	defer proxy.Stop()

	resp, err := http.Get("http://localhost:61901/v1.15/containers/foo/json")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `{"Id":"doesnt_matter", "Name":"doesnt_matter"}`, string(body))
}

func TestStart(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61902",
	}

	go proxy.Serve()
	defer proxy.Stop()

	resp, err := http.Post("http://localhost:61902/v1.15/containers/foo/start", "", nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestListAll_NoContainersCreated(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61903",
	}

	go proxy.Serve()
	defer proxy.Stop()

	resp, err := http.Get("http://localhost:61903/v1.15/containers/json")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "[]", string(body))
}

func TestListAll_WithContainersCreated(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61904",
		containerRequests: []ContainerRequest{
			{Name: "foo"},
			{Name: "bar"},
		},
	}

	go proxy.Serve()
	defer proxy.Stop()

	resp, err := http.Get("http://localhost:61904/v1.15/containers/json")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	var cs []dockerclient.Container
	json.Unmarshal(body, &cs)

	assert.Len(t, cs, 2)
	assert.Equal(t, []string{"foo"}, cs[0].Names)
	assert.Equal(t, []string{"bar"}, cs[1].Names)
}
