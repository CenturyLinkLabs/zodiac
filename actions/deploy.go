package actions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/discovery"
	log "github.com/Sirupsen/logrus"
	"github.com/k0kubun/pp"
	"github.com/samalba/dockerclient"
)

type AttemptedContainer struct {
	Name          string
	CreateOptions []byte
	Config        dockerclient.ContainerConfig
}

func Deploy(c discovery.Cluster) (prettycli.Output, error) {
	attemptedContainers := make([]AttemptedContainer, 0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bs, _ := ioutil.ReadAll(r.Body)
		log.Infof("Got: %s %s, Body: %s", r.Method, r.URL.String(), string(bs))

		if r.URL.Path == "/v1.15/containers/beefbeef/json" {
			fmt.Fprintf(w, `{"Id": "beefbeef", "Name": "/hangry_welch"}`)
		} else if r.URL.Path == "/v1.15/containers/beefbeef/start" {
			w.WriteHeader(http.StatusNoContent)
		} else if strings.HasPrefix(r.URL.String(), "/v1.15/containers/create") {
			name := r.URL.Query()["name"][0]
			attemptedContainers = append(attemptedContainers, AttemptedContainer{
				CreateOptions: bs,
				Name:          name,
			})

			fmt.Fprintf(w, `{ "Id":"beefbeef", "Warnings":[] }`)
		} else {
			r.URL.Host = "10.134.246.158:2375"
			r.URL.Scheme = "http"
			log.Infof("Proxied %s", r.URL.String())
			req, err := http.NewRequest(r.Method, r.URL.String(), nil)
			c := http.Client{}
			resp, err := c.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			respBody, _ := ioutil.ReadAll(resp.Body)
			fmt.Fprintf(w, string(respBody))
		}
	})

	fmt.Println("So uhhhh, press any key when you're deployed... Or Ctrl-C if you want to abort...")
	laddr, _ := net.ResolveTCPAddr("tcp", "localhost:31981")
	listener, _ := net.ListenTCP("tcp", laddr)
	go http.Serve(listener, handler)
	defer listener.Close()

	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	for _, ac := range attemptedContainers {
		var cc dockerclient.ContainerConfig
		_ = json.Unmarshal(ac.CreateOptions, &cc)
		ac.Config = cc
	}

	client, err := dockerclient.NewDockerClient("tcp://10.134.246.158:2375", nil)
	if err != nil {
		log.Fatal(err)
	}
	pp.Print(attemptedContainers)

	for _, ac := range attemptedContainers {
		if ac.Config.Labels == nil {
			ac.Config.Labels = make(map[string]string)
		}
		ac.Config.Labels["zodiac"] = "strike"
		id, err := client.CreateContainer(&ac.Config, ac.Name)
		if err != nil {
			log.Fatal("problem creating", err)
		}
		log.Infof("%s created as %s", ac.Name, id)
		if err := client.StartContainer(id, &dockerclient.HostConfig{}); err != nil {
			log.Fatal("problem starting", err)
		}
		log.Infof("%s started")
	}

	return prettycli.PlainOutput{"Would deploy..."}, nil
}
