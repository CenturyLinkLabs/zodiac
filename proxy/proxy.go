package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/CenturyLinkLabs/zodiac/endpoint"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/samalba/dockerclient"
)

type ContainerRequest struct {
	Name          string
	CreateOptions []byte
}

type ProxyFactory func(string, endpoint.Endpoint, bool) Proxy

type Proxy interface {
	Serve() error
	Stop() error
	GetRequests() ([]ContainerRequest, error)
}

type HTTPProxy struct {
	address            string
	containerRequests  []ContainerRequest
	listener           *net.TCPListener
	errors             []error
	imageInspectsCount map[string]int
	endpoint           endpoint.Endpoint
	noBuild            bool
}

func NewHTTPProxy(listenAt string, endpoint endpoint.Endpoint, noBuild bool) Proxy {
	return &HTTPProxy{
		address:  listenAt,
		endpoint: endpoint,
		noBuild:  noBuild,
	}
}

func (p *HTTPProxy) Serve() error {
	r := mux.NewRouter()
	r.Path("/v1.18/containers/create").Methods("POST").HandlerFunc(p.create)
	r.Path("/v1.18/containers/{id}/json").Methods("GET").HandlerFunc(p.inspect)
	r.Path("/v1.18/containers/{id}/start").Methods("POST").HandlerFunc(p.start)
	r.Path("/v1.18/containers/json").Methods("GET").HandlerFunc(p.listAll)
	r.Path("/v1.18/images/{id:.*}/json").Methods("GET").HandlerFunc(p.inspectImage)
	r.Path("/v1.18/build").Methods("POST").HandlerFunc(p.build)
	r.Path("/v1.18/images/create").Methods("POST").HandlerFunc(p.createImage)
	r.Path("/{rest:.*}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Unhandled request to: %s\n\n", r.URL)
	})

	laddr, _ := net.ResolveTCPAddr("tcp", p.address)
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

	fmt.Fprintf(w, `{"Id":"doesnt_matter", "Warnings":[]}`)
}

func (p *HTTPProxy) inspect(w http.ResponseWriter, r *http.Request) {
	log.Infof("INSPECT request to %s", r.URL)
	containerInfo := dockerclient.ContainerInfo{
		Id:   "id_doesnt_matter",
		Name: "name_doesnt_matter",
		Config: &dockerclient.ContainerConfig{
			Labels: map[string]string{
				"com.docker.compose.container-number": "1",
			},
		},
	}
	j, err := json.Marshal(containerInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(j)
}

func (p *HTTPProxy) start(w http.ResponseWriter, r *http.Request) {
	log.Infof("START request to %s", r.URL)
	w.WriteHeader(204)
}

func (p *HTTPProxy) listAll(w http.ResponseWriter, r *http.Request) {
	log.Infof("LIST ALL request to %s", r.URL)

	filters := r.URL.Query()["filters"]

	name := filteredServiceName(filters)

	containers := []dockerclient.Container{}

	for _, req := range p.containerRequests {
		if (filters == nil) || (extractReqName(req.Name) == name) {
			container := dockerclient.Container{
				Id:    "abc123",
				Image: "doesntmatter",
				Names: []string{req.Name},
				Labels: map[string]string{
					"com.docker.compose.container-number": "1",
				},
			}
			containers = append(containers, container)
		}
	}

	j, err := json.Marshal(containers)
	if err != nil {
		p.errors = append(p.errors, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(j)
}

func (p *HTTPProxy) inspectImage(w http.ResponseWriter, r *http.Request) {
	log.Infof("IMAGE INSPECT request to %s", r.URL)
	id := mux.Vars(r)["id"]
	if p.imageInspectsCount == nil {
		p.imageInspectsCount = make(map[string]int)
	}
	p.imageInspectsCount[id] = p.imageInspectsCount[id] + 1

	img := &dockerclient.ImageInfo{}
	jres, err := json.Marshal(img)
	if err != nil {
		p.errors = append(p.errors, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if p.noBuild || p.imageInspectsCount[id] > 1 {
		fmt.Fprintf(w, string(jres))
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No such image: doesnt_matter")
	}
}

func (p *HTTPProxy) createImage(w http.ResponseWriter, r *http.Request) {
	log.Infof("IMAGE CREATE REQUEST to %s", r.URL)

	fmt.Fprintf(w, `{}`)
}

func (p *HTTPProxy) build(w http.ResponseWriter, r *http.Request) {
	log.Infof("BUILD REQUEST to %s", r.URL)

	if !p.noBuild {
		err := p.endpoint.BuildImage(r.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Fprintf(w, `{"stream":"Successfully built abc123\n"}`)
}

func extractReqName(reqName string) string {
	parts := strings.Split(reqName, "_")
	newParts := parts[1 : len(parts)-1]
	return strings.Join(newParts, "_")
}

func filteredServiceName(filters []string) string {
	if filters != nil {
		filter := filters[0]
		x := map[string][]string{}
		json.Unmarshal([]byte(filter), &x)
		for _, item := range x["label"] {
			if strings.HasPrefix(item, "com.docker.compose.service") {
				return strings.SplitAfterN(item, "=", 2)[1]
			}
		}
	}
	return ""
}
