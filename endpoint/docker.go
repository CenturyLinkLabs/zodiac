package endpoint

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/user"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

func init() {
}

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

type EndpointOptions struct {
	Host      string
	TLS       bool
	TLSVerify bool
	TLSCaCert string
	TLSCert   string
	TLSKey    string
}

func (eo EndpointOptions) tlsCaCert() string {
	return resolveHomeDirectory(eo.TLSCaCert)
}

func (eo EndpointOptions) tlsCert() string {
	return resolveHomeDirectory(eo.TLSCert)
}

func (eo EndpointOptions) tlsKey() string {
	return resolveHomeDirectory(eo.TLSKey)
}

func resolveHomeDirectory(path string) string {
	if strings.Contains(path, "~") {
		usr, _ := user.Current()
		dir := usr.HomeDir
		return strings.Replace(path, "~", dir, 1)
	}

	return path
}

func getTlsConfig(endpointOpts EndpointOptions) (*tls.Config, error) {
	var tlsConfig *tls.Config

	if endpointOpts.TLS {

		tlsConfig = &tls.Config{
			InsecureSkipVerify: !endpointOpts.TLSVerify,
		}

		if endpointOpts.tlsCert() != "" && endpointOpts.tlsKey() != "" {

			cert, err := tls.LoadX509KeyPair(endpointOpts.tlsCert(), endpointOpts.tlsKey())
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		// Load CA cert
		if endpointOpts.tlsCaCert() != "" {

			caCert, err := ioutil.ReadFile(endpointOpts.tlsCaCert())
			if err != nil {
				return nil, err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsConfig.RootCAs = caCertPool
		}
	}
	return tlsConfig, nil
}

type EndpointFactory func(EndpointOptions) (Endpoint, error)

type Endpoint interface {
	Version() (string, error)
	Name() string
	Host() string
	DoRequest(*http.Request) (*http.Response, error)
	ResolveImage(string) (string, error)
	StartContainer(name string, cc ContainerConfig) error
	InspectContainer(name string) (*dockerclient.ContainerInfo, error)
	RemoveContainer(name string) error
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

func translateContainerConfig(cc ContainerConfig) (dockerclient.ContainerConfig, error) {
	var dcc dockerclient.ContainerConfig

	j, err := json.Marshal(cc)
	if err != nil {
		return dcc, err
	}

	err = json.Unmarshal(j, &dcc)
	return dcc, err
}

type ContainerConfig struct {
	dockerclient.ContainerConfig
	Tty        FromStringOrBool
	OpenStdin  FromStringOrBool
	Entrypoint FromStringOrStringSlice
	// This is used only by the create command
	HostConfig HostConfig
}

type HostConfig struct {
	dockerclient.HostConfig
	Privileged     FromStringOrBool
	ReadonlyRootfs FromStringOrBool
}

type FromStringOrBool struct {
	Value bool
}

func (s *FromStringOrBool) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		var str string
		if err := json.Unmarshal(value, &str); err != nil {
			return err
		}
		s.Value = (strings.ToLower(str) == "true")
		return nil
	}

	return json.Unmarshal(value, &s.Value)
}

func (s FromStringOrBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}

type FromStringOrStringSlice struct {
	Value []string
}

func (s *FromStringOrStringSlice) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		var str string
		if err := json.Unmarshal(value, &str); err != nil {
			return err
		}
		s.Value = strings.Split(str, " ")
		return nil
	}

	return json.Unmarshal(value, &s.Value)
}

func (s FromStringOrStringSlice) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}
