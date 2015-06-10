package fakeengine

import (
	"net/http"
	"net/http/httptest"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/samalba/dockerclient"
)

func NewServer() *httptest.Server {
	r := mux.NewRouter()
	baseURL := "/" + dockerclient.APIVersion
	r.HandleFunc(baseURL+"/version", handlerGetVersion).Methods("GET")
	r.HandleFunc(baseURL+"/images/{name}/json", handleInspectImage).Methods("GET")
	r.HandleFunc(baseURL+"/containers/create", handleCreateContainer).Methods("POST")
	r.HandleFunc(baseURL+"/containers/{id}/start", handleStartContainer).Methods("POST")
	r.HandleFunc("/{rest:.*}", catchAll)
	return httptest.NewServer(handlerAccessLog(r))
}

func handlerAccessLog(handler http.Handler) http.Handler {
	logHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Debugf(`%s "%s %s"`, r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(logHandler)
}

func writeHeaders(w http.ResponseWriter, code int) {
	h := w.Header()
	h.Add("Content-Type", "application/json")
	w.WriteHeader(code)
}

func handlerGetVersion(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200)
	body := `{
		"Version": "1.6.0",
		"Os": "linux",
		"KernelVersion": "3.18.5-tinycore64",
		"GoVersion": "go1.4.1",
		"GitCommit": "a8a31ef",
		"Arch": "amd64",
		"ApiVersion": "1.18"
	}`
	w.Write([]byte(body))
}

func handleInspectImage(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200)
	body := `{
		"Id": "abc123"
	}`
	w.Write([]byte(body))
}

func handleCreateContainer(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 201)
	body := `{
		"Id":"e90e34656806",
		"Warnings":[]
	}`
	w.Write([]byte(body))
}

func handleStartContainer(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 204)
}

func catchAll(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200)
	w.Write([]byte("[]"))
}
