package actions

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

var (
	endpointFactory EndpointFactory
)

func init() {
	endpointFactory = func(endpointOpts EndpointOptions) (Endpoint, error) {

		tlsConfig := &tls.Config{
			InsecureSkipVerify: !endpointOpts.TLSVerify,
		}

		if endpointOpts.TLSCert != "" && endpointOpts.TLSKey != "" {
			cert, err := tls.LoadX509KeyPair(endpointOpts.TLSCert, endpointOpts.TLSKey)
			if err != nil {
				log.Fatal(err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		// Load CA cert
		if endpointOpts.TLSCaCert != "" {
			caCert, err := ioutil.ReadFile(endpointOpts.TLSCaCert)
			if err != nil {
				log.Fatal(err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsConfig.RootCAs = caCertPool
		}

		c, err := dockerclient.NewDockerClient(endpointOpts.Host, tlsConfig)
		if err != nil {
			return nil, err
		}

		return &DockerEndpoint{url: endpointOpts.Host, client: c}, nil
	}
}

type EndpointFactory func(EndpointOptions) (Endpoint, error)

type Endpoint interface {
	Version() (string, error)
	Name() string
	Host() string
	ResolveImage(string) (string, error)
	StartContainer(name string, cc ContainerConfig) error
	InspectContainer(name string) (*dockerclient.ContainerInfo, error)
	RemoveContainer(name string) error
}

type DockerEndpoint struct {
	url    string
	client dockerclient.Client
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
	id, err := e.client.CreateContainer(&dcc, name)
	if err != nil {
		log.Fatalf("Problem creating container: ", err)
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

func (e *DockerEndpoint) InspectContainer(name string) (*dockerclient.ContainerInfo, error) {
	return e.client.InspectContainer(name)
}

func (e *DockerEndpoint) RemoveContainer(name string) error {
	//TODO: be more graceful
	return e.client.RemoveContainer(name, true, false)
}

func translateContainerConfig(cc ContainerConfig) (dockerclient.ContainerConfig, error) {
	var dcc dockerclient.ContainerConfig

	j, err := json.Marshal(cc)
	if err != nil {
		return dcc, err
	}

	err = json.Unmarshal(j, &dcc)
	return dcc, err
}
