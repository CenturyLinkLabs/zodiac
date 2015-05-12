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
	return httptest.NewServer(handlerAccessLog(r))
}

func handlerAccessLog(handler http.Handler) http.Handler {
	logHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Debugf(`%s "%s %s"`, r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(logHandler)
}

func writeHeaders(w http.ResponseWriter, code int, jobName string) {
	h := w.Header()
	h.Add("Content-Type", "application/json")
	if jobName != "" {
		h.Add("Job-Name", jobName)
	}
	w.WriteHeader(code)
}

func handlerGetVersion(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "version")
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
