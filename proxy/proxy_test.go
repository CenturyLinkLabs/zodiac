package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
)

func TestGetRequests_WithNoErrors(t *testing.T) {
	proxy := HTTPProxy{
		containerRequests: []ContainerRequest{
			{Name: "zodiac_foo_1"},
			{Name: "zodiac_bar_1"},
		},
	}

	reqs, err := proxy.GetRequests()
	assert.NoError(t, err)
	assert.Len(t, reqs, 2)
	assert.Equal(t, "zodiac_foo_1", reqs[0].Name)
	assert.Equal(t, "zodiac_bar_1", reqs[1].Name)
}

func TestGetRequests_WithErrors(t *testing.T) {
	proxy := HTTPProxy{
		containerRequests: []ContainerRequest{
			{Name: "zodiac_foo_1"},
			{Name: "zodiac_bar_1"},
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

	go proxy.Serve("test.local", false)
	defer proxy.Stop()

	resp, err := http.Post("http://localhost:61900/v1.18/containers/create?name=foo", "", strings.NewReader("bar"))

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

	go proxy.Serve("test.local", false)
	defer proxy.Stop()

	resp, err := http.Get("http://localhost:61901/v1.18/containers/foo/json")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)

	expectedContainerInfo := dockerclient.ContainerInfo{
		Id:   "id_doesnt_matter",
		Name: "name_doesnt_matter",
		Config: &dockerclient.ContainerConfig{
			Labels: map[string]string{
				"com.docker.compose.container-number": "1",
			},
		},
	}
	j, err := json.Marshal(expectedContainerInfo)
	assert.Equal(t, string(j), string(body))
}

func TestStart(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61902",
	}

	go proxy.Serve("test.local", false)
	defer proxy.Stop()

	resp, err := http.Post("http://localhost:61902/v1.18/containers/foo/start", "", nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestListAll_NoContainersCreated(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61903",
	}

	go proxy.Serve("test.local", false)
	defer proxy.Stop()

	resp, err := http.Get("http://localhost:61903/v1.18/containers/json")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "[]", string(body))
}

func TestListAll_WithContainersCreated(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61904",
		containerRequests: []ContainerRequest{
			{Name: "zodiac_foo_1"},
			{Name: "zodiac_bar_1"},
		},
	}

	go proxy.Serve("test.local", false)
	defer proxy.Stop()

	resp, err := http.Get("http://localhost:61904/v1.18/containers/json")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	var cs []dockerclient.Container
	json.Unmarshal(body, &cs)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, cs, 2)
	assert.Equal(t, []string{"zodiac_foo_1"}, cs[0].Names)
	assert.Equal(t, []string{"zodiac_bar_1"}, cs[1].Names)
}

func TestListAll_WithMatchingFilteredRequest(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61905",
		containerRequests: []ContainerRequest{
			{Name: "zodiac_fiz_biz_1"},
			{Name: "zodiac_bar_1"},
		},
	}

	go proxy.Serve("test.local", false)
	defer proxy.Stop()

	query := url.QueryEscape(`{"label": ["com.docker.compose.project=zodiac", "com.docker.compose.service=fiz_biz", "com.docker.compose.oneoff=False"]}`)

	resp, err := http.Get(fmt.Sprintf("http://localhost:61905/v1.18/containers/json?filters=%s", query))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	var cs []dockerclient.Container
	json.Unmarshal(body, &cs)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, cs, 1)
	assert.Equal(t, []string{"zodiac_fiz_biz_1"}, cs[0].Names)
}

func TestListAll_WithNonMatchingFilteredRequest(t *testing.T) {
	proxy := HTTPProxy{
		address: "localhost:61906",
		containerRequests: []ContainerRequest{
			{Name: "zodiac_foo_1"},
			{Name: "zodiac_bar_1"},
		},
	}

	go proxy.Serve("test.local", false)
	defer proxy.Stop()

	query := url.QueryEscape(`{"label": ["com.docker.compose.project=zodiac", "com.docker.compose.service=DOES_NOT_MATCH", "com.docker.compose.oneoff=False"]}`)

	resp, err := http.Get(fmt.Sprintf("http://localhost:61906/v1.18/containers/json?filters=%s", query))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	var cs []dockerclient.Container
	json.Unmarshal(body, &cs)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, cs, 0)
}
