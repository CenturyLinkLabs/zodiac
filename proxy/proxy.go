package proxy

import (
	"net"
	"net/http"
	"io/ioutil"
	"fmt"

	"github.com/samalba/dockerclient"
	"github.com/gorilla/mux"
	log "github.com/Sirupsen/logrus"
	"github.com/CenturyLinkLabs/zodiac/cluster"
)

type ContainerRequest struct {
	Name          string
	CreateOptions []byte
	Config        dockerclient.ContainerConfig
}

type Proxy interface {
	Serve(cluster.Endpoint) error
	Stop() error
	DrainRequests() []*ContainerRequest
}

type HTTPProxy struct {
	//containerRequests []*ContainerRequest
	listener *net.TCPListener
}

func NewHTTPProxy(listenAt string) HTTPProxy {
	// TODO: Implement. Don't have this do anything resource-intensive since it runs at init.

	return HTTPProxy{}
}

func (p *HTTPProxy) Serve(endpoint cluster.Endpoint) error {
	r := mux.NewRouter()
	r.Path("/v1.15/containers/create").Methods("POST").HandlerFunc(create)
	r.Path("/{rest:.*}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		catchAll(endpoint, w, r)
	})

	// TODO: port should be configurable
	laddr, _ := net.ResolveTCPAddr("tcp", "localhost:3000")
	listener, _ := net.ListenTCP("tcp", laddr)
	p.listener = listener
	return http.Serve(listener, r)
}

func catchAll(endpoint cluster.Endpoint, w http.ResponseWriter, r *http.Request) {
	log.Debugf("unhandled request to %s", r.URL)


	r.URL.Host = endpoint.Name()
	// TODO: we should give the scheme some thought
	r.URL.Scheme = "http"
	// log.Infof("Proxied %s", r.URL.String())
	req, err := http.NewRequest(r.Method, r.URL.String(), nil)
	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// TODO, is there a better way?
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Fprintf(w, string(respBody))
}

func (p *HTTPProxy) Stop() error {
	p.listener.Close()
	return nil
}

func (p *HTTPProxy) DrainRequests() []*ContainerRequest {
	// TODO implement, and don't forget that this is supposed to 'drain', as in
	// remove the saved requests that this instance has.
	return make([]*ContainerRequest, 0)
}

func create(w http.ResponseWriter, r *http.Request) {
	log.Info("Creating")
}
