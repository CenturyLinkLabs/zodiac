package endpoint

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

func NewEndpoint(endpointOpts EndpointOptions) (Endpoint, error) {
	tlsConfig, err := getTlsConfig(endpointOpts)
	if err != nil {
		return nil, err
	}

	c, err := dockerclient.NewDockerClient(endpointOpts.Host, tlsConfig)
	if err != nil {
		return nil, err
	}

	return &DockerEndpoint{
		url:       endpointOpts.Host,
		client:    c,
		tlsConfig: tlsConfig,
	}, nil
}

type DockerEndpoint struct {
	url       string
	client    dockerclient.Client
	tlsConfig *tls.Config
}

// TODO: can we ditch this? Should always have it on the client
func (e *DockerEndpoint) Host() string {
	url, _ := url.Parse(e.url)
	return url.Host
}

func (e *DockerEndpoint) Version() (string, error) {
	v, err := e.client.Version()
	if err != nil {
		return "", err
	}

	return v.Version, nil
}

// TODO: can we ditch this? Should always have it on the client
func (e *DockerEndpoint) Name() string {
	return e.url
}

func (e *DockerEndpoint) StartContainer(name string, cc ContainerConfig) error {
	dcc, _ := translateContainerConfig(cc)
	var id string
	for {
		cid, err := e.client.CreateContainer(&dcc, name)
		if err == nil {
			id = cid
			break
		} else {
			log.Infof("Problem creating container: ", err)
			log.Infof("Retrying create...")
			time.Sleep(5 * time.Second)
		}
	}

	log.Infof("%s created as %s", name, id)

	if err := e.client.StartContainer(id, nil); err != nil {
		log.Fatal("problem starting: ", err)
	}
	return nil
}

func (e *DockerEndpoint) ResolveImage(name string) (string, error) {
	imageInfo, err := e.client.InspectImage(name)
	if err != nil {

		if err == dockerclient.ErrNotFound {
			if err := e.client.PullImage(name, nil); err != nil {
				return "", err
			}
			imageInfo, err = e.client.InspectImage(name)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	return imageInfo.Id, nil
}

func (e *DockerEndpoint) BuildImage(buildContext io.Reader, svcName string) error {
	scheme := "https"
	if e.tlsConfig == nil {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s/%s/build?pull=False&nocache=False&q=False&t=%s&forcerm=False&rm=True", scheme, e.Host(), dockerclient.APIVersion, svcName)
	req, err := http.NewRequest("POST", url, buildContext)
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/tar")
	tr := &http.Transport{
		TLSClientConfig: e.tlsConfig,
	}
	c := http.Client{Transport: tr}
	resp, err := c.Do(req)

	if err != nil {
		return err
	}

	// NOTE: this is here to make sure it finishes
	_, err = ioutil.ReadAll(resp.Body)
	return err
}

func (e *DockerEndpoint) InspectContainer(name string) (*dockerclient.ContainerInfo, error) {
	return e.client.InspectContainer(name)
}

func (e *DockerEndpoint) RemoveContainer(name string) error {
	//TODO: be more graceful
	return e.client.RemoveContainer(name, true, false)
}
