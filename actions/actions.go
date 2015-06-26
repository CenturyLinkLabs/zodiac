package actions

import (
	"encoding/json"
	"fmt"
	"os/user"
	"strings"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/composer"
	"github.com/CenturyLinkLabs/zodiac/proxy"
	"github.com/samalba/dockerclient"
)

const (
	ProxyAddress  = "localhost:61908"
	BasicDateTime = "2006-01-02 15:04:05"
)

var (
	DefaultProxy    proxy.Proxy
	DefaultComposer composer.Composer
)

func init() {
	DefaultProxy = proxy.NewHTTPProxy(ProxyAddress)
	DefaultComposer = composer.NewExecComposer(ProxyAddress)
}

type Options struct {
	Args            []string
	Flags           map[string]string
	EndpointOptions EndpointOptions
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

type Zodiaction func(Options) (prettycli.Output, error)

type DeploymentManifests []DeploymentManifest

type DeploymentManifest struct {
	Services   []Service
	DeployedAt string
	Message    string
}

type Service struct {
	Name            string
	ContainerConfig ContainerConfig
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

func collectRequests(options Options) ([]proxy.ContainerRequest, error) {
	go DefaultProxy.Serve()
	defer DefaultProxy.Stop()

	if err := DefaultComposer.Run(options.Flags); err != nil {
		return nil, err
	}

	return DefaultProxy.GetRequests()
}

func startServices(services []Service, manifests DeploymentManifests, endpoint Endpoint) error {
	manifestsBlob, err := json.Marshal(manifests)
	if err != nil {
		return err
	}

	for _, svc := range services {
		if svc.ContainerConfig.Labels == nil {
			svc.ContainerConfig.Labels = make(map[string]string)
		}
		svc.ContainerConfig.Labels["zodiacManifest"] = string(manifestsBlob)

		fmt.Printf("Creating %s\n", svc.Name)

		if err := endpoint.StartContainer(svc.Name, svc.ContainerConfig); err != nil {
			return err
		}
	}

	return nil
}

func getDeploymentManifests(reqs []proxy.ContainerRequest, endpoint Endpoint) (DeploymentManifests, error) {
	var manifests DeploymentManifests
	var inspectError error

	for _, req := range reqs {
		ci, err := endpoint.InspectContainer(req.Name)
		inspectError = err
		if err != nil {
			continue
		}

		if err := json.Unmarshal([]byte(ci.Config.Labels["zodiacManifest"]), &manifests); err != nil {
			return nil, err
		}

		break
	}

	if inspectError != nil {
		return nil, inspectError
	}

	return manifests, nil
}
