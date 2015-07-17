package endpoint

import (
	"crypto/tls"
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
		useTLS:    endpointOpts.TLS,
	}, nil
}

type DockerEndpoint struct {
	url       string
	client    dockerclient.Client
	tlsConfig *tls.Config
	useTLS    bool
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

func (e *DockerEndpoint) DoRequest(r *http.Request) (*http.Response, error) {
	r.URL.Host = e.Host()
	r.URL.Scheme = "http"
	if e.useTLS {
		r.URL.Scheme = "https"
	}
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/tar")
	tr := &http.Transport{
		TLSClientConfig: e.tlsConfig,
	}
	c := http.Client{Transport: tr}
	return c.Do(req)
}

func (e *DockerEndpoint) InspectContainer(name string) (*dockerclient.ContainerInfo, error) {
	return e.client.InspectContainer(name)
}

func (e *DockerEndpoint) RemoveContainer(name string) error {
	//TODO: be more graceful
	return e.client.RemoveContainer(name, true, false)
}
