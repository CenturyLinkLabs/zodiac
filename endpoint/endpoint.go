package endpoint

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/user"
	"strings"

	"github.com/samalba/dockerclient"
)

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
