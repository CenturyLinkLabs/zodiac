package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/CenturyLinkLabs/zodiac/cluster"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/samalba/dockerclient"
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

func NewHTTPProxy(listenAt string) *HTTPProxy {
	return &HTTPProxy{}
}

func (p *HTTPProxy) Serve(endpoint cluster.Endpoint) error {
	r := mux.NewRouter()
	r.Path("/v1.15/containers/create").Methods("POST").HandlerFunc(p.create)
	r.Path("/v1.15/containers/{id}/json").Methods("GET").HandlerFunc(p.inspect)
	r.Path("/v1.15/containers/{id}/start").Methods("POST").HandlerFunc(p.start)
	r.Path("/v1.15/containers/json").Methods("GET").HandlerFunc(p.listAll)

	// TODO: port should be configurable
	laddr, _ := net.ResolveTCPAddr("tcp", "localhost:3000")
	listener, _ := net.ListenTCP("tcp", laddr)
	p.listener = listener
	return http.Serve(listener, r)
}

func (p *HTTPProxy) Stop() error {
	p.listener.Close()
	return nil
}

func (p *HTTPProxy) DrainRequests() []cluster.ContainerRequest {
	// TODO maybe drain errors through this method as well
	// TODO implement, and don't forget that this is supposed to 'drain', as in
	// remove the saved requests that this instance has.
	return p.containerRequests
}

func (p *HTTPProxy) create(w http.ResponseWriter, r *http.Request) {
	log.Infof("CREATE request to %s", r.URL)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		//TODO: return errors to compose
	}

	name := r.URL.Query()["name"][0]

	req := cluster.ContainerRequest{
		Name:          name,
		CreateOptions: body,
	}

	p.containerRequests = append(p.containerRequests, req)

	fmt.Fprintf(w, `{ "Id":"doesnt_matter", "Warnings":[] }`)
}

func (p *HTTPProxy) inspect(w http.ResponseWriter, r *http.Request) {
	log.Infof("INSPECT request to %s", r.URL)
	fmt.Fprintf(w, `{"Id": "doesnt_matter", "Name": "doesnt_matter"}`)
}

func (p *HTTPProxy) start(w http.ResponseWriter, r *http.Request) {
	log.Infof("START request to %s", r.URL)
	w.WriteHeader(204)
}

func (p *HTTPProxy) listAll(w http.ResponseWriter, r *http.Request) {
	log.Infof("LIST ALL request to %s", r.URL)

	containers := []dockerclient.Container{}

	for _, req := range p.containerRequests {
		containerInfo := dockerclient.Container{
			Image: req.Config.Image,
			Names: []string{req.Name},
		}
		containers = append(containers, containerInfo)
	}

	j, _ := json.Marshal(containers)
	w.Write(j)
}
