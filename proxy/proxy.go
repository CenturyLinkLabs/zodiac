package proxy

import "github.com/samalba/dockerclient"

type ContainerRequest struct {
	Name          string
	CreateOptions []byte
	Config        dockerclient.ContainerConfig
}

type Proxy interface {
	Serve() error
	Stop() error
	DrainRequests() []*ContainerRequest
}

type HTTPProxy struct {
	//containerRequests []*ContainerRequest
}

func NewHTTPProxy(listenAt string) HTTPProxy {
	// TODO: Implement. Don't have this do anything resource-intensive since it runs at init.
	return HTTPProxy{}
}

func (p *HTTPProxy) Serve() error {
	return nil
}

func (p *HTTPProxy) Stop() error {
	return nil
}

func (p *HTTPProxy) DrainRequests() []*ContainerRequest {
	// TODO implement, and don't forget that this is supposed to 'drain', as in
	// remove the saved requests that this instance has.
	return make([]*ContainerRequest, 0)
}
