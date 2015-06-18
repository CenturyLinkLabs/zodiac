package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/samalba/dockerclient"
)

type ContainerRequest struct {
	Name          string
	CreateOptions []byte
	Config        dockerclient.ContainerConfig
}

type Proxy interface {
	Serve() error
	Stop() error
	GetRequests() ([]ContainerRequest, error)
}

type HTTPProxy struct {
	containerRequests []ContainerRequest
	listener          *net.TCPListener
	errors            []error
}

func NewHTTPProxy(listenAt string) *HTTPProxy {
	return &HTTPProxy{}
}

func (p *HTTPProxy) Serve() error {
	r := mux.NewRouter()
	r.Path("/v1.15/containers/create").Methods("POST").HandlerFunc(p.create)
	r.Path("/v1.15/containers/{id}/json").Methods("GET").HandlerFunc(p.inspect)
	r.Path("/v1.15/containers/{id}/start").Methods("POST").HandlerFunc(p.start)
	r.Path("/v1.15/containers/json").Methods("GET").HandlerFunc(p.listAll)

	laddr, _ := net.ResolveTCPAddr("tcp", "localhost:61908")
	listener, _ := net.ListenTCP("tcp", laddr)
	p.listener = listener
	return http.Serve(listener, r)
}

func (p *HTTPProxy) Stop() error {
	p.listener.Close()
	return nil
}

func (p *HTTPProxy) GetRequests() ([]ContainerRequest, error) {
	if len(p.errors) > 0 {
		// TODO: collect errors?
		return nil, errors.New("Error parsing compose template")
	}

	return p.containerRequests, nil
}

func (p *HTTPProxy) create(w http.ResponseWriter, r *http.Request) {
	log.Infof("CREATE request to %s", r.URL)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		p.errors = append(p.errors, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	name := r.URL.Query()["name"][0]

	req := ContainerRequest{
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

	j, err := json.Marshal(containers)
	if err != nil {
		p.errors = append(p.errors, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(j)
}
