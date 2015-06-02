package proxy

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/CenturyLinkLabs/zodiac/cluster"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type Proxy interface {
	Serve(cluster.Endpoint) error
	Stop() error
	DrainRequests() []cluster.ContainerRequest
}

type HTTPProxy struct {
	containerRequests []cluster.ContainerRequest
	listener          *net.TCPListener
}

func NewHTTPProxy(listenAt string) HTTPProxy {
	// TODO: Implement. Don't have this do anything resource-intensive since it runs at init.

	return HTTPProxy{
		containerRequests: make([]cluster.ContainerRequest, 0),
	}

}

func (p *HTTPProxy) Serve(endpoint cluster.Endpoint) error {
	r := mux.NewRouter()
	r.Path("/v1.15/containers/create").Methods("POST").HandlerFunc(p.create)
	r.Path("/v1.15/containers/{id}/json").Methods("GET").HandlerFunc(p.inspect)
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
	log.Infof("unhandled request to %s", r.URL)

	log.Debugf("Logging request HEADERs")
	for k, v := range r.Header {
		log.Debugf(k, v)
	}

	r.URL.Host = endpoint.Name()
	// TODO: we should give the scheme some thought
	r.URL.Scheme = "http"
	// log.Infof("Proxied %s", r.URL.String())
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	req.Header.Set("content-type", "application/json")
	c := http.Client{}
	resp, err := c.Do(req)
	resp.Header.Set("content-type", "application/json")
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("Logging RESPONSE HEADERs")
	for k, v := range resp.Header {
		log.Debugf(k, v)
	}

	// TODO, is there a better way?
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Fprintf(w, string(respBody))
}

func (p *HTTPProxy) Stop() error {
	p.listener.Close()
	return nil
}

func (p *HTTPProxy) DrainRequests() []cluster.ContainerRequest {
	// TODO implement, and don't forget that this is supposed to 'drain', as in
	// remove the saved requests that this instance has.
	return p.containerRequests
}

func (p *HTTPProxy) create(w http.ResponseWriter, r *http.Request) {
	log.Infof("CREATE request to %s", r.URL)
	//log.Infof("CREATE form %s", r.Form)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		//TODO: return errors to compose
	}

	//TODO: not all requests have a name if the service is being restarted
	req := cluster.ContainerRequest{
		Name:          r.URL.Query()["name"][0],
		CreateOptions: body,
	}

	p.containerRequests = append(p.containerRequests, req)

	fmt.Fprintf(w, `{ "Id":"foo", "Warnings":[] }`)
}

func (p *HTTPProxy) inspect(w http.ResponseWriter, r *http.Request) {
	log.Infof("INSPECT request to %s", r.URL)
	fmt.Fprintf(w, `{"Id": "foo", "Name": "/hangry_welch"}`)
}
