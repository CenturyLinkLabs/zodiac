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

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/discovery"
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

var RegistryClient = registry.NewClient()

type AttemptedContainer struct {
	Name          string
	CreateOptions []byte
	Config        dockerclient.ContainerConfig
}

type DeploymentManifest map[string]string

func Deploy(c discovery.Cluster) (prettycli.Output, error) {
	attemptedContainers := make([]*AttemptedContainer, 0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bs, _ := ioutil.ReadAll(r.Body)
		log.Infof("Got: %s %s, Body: %s", r.Method, r.URL.String(), string(bs))

		if r.URL.Path == "/v1.15/containers/beefbeef/json" {
			fmt.Fprintf(w, `{"Id": "beefbeef", "Name": "/hangry_welch"}`)
		} else if r.URL.Path == "/v1.15/containers/beefbeef/start" {
			w.WriteHeader(http.StatusNoContent)
		} else if strings.HasPrefix(r.URL.String(), "/v1.15/containers/create") {
			nameParam := r.URL.Query()["name"]
			if len(nameParam) != 1 {
				log.Fatal("There was a problem with the container creation request")
			}

			name := nameParam[0]
			attemptedContainers = append(attemptedContainers, &AttemptedContainer{
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
		if err := json.Unmarshal(ac.CreateOptions, &cc); err != nil {
			log.Fatalf("Error unmarshalling request JSON for '%s': %s", ac.Name, err.Error())
		}
		ac.Config = cc
	}

	client, err := dockerclient.NewDockerClient("tcp://10.134.246.158:2375", nil)
	if err != nil {
		log.Fatal(err)
	}

	manifest := make(DeploymentManifest)
	for _, ac := range attemptedContainers {
		sha, err := resolveImage(ac.Config.Image)
		if err != nil {
			log.Fatalf("the image '%s' couldn't be found: %s", ac.Config.Image, err.Error())
		}
		manifest[ac.Name] = sha
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		log.Fatal("there was a problem building the deployment manafest: ", err)
	}

	for _, ac := range attemptedContainers {
		// TODO Why? Change the way the Config is instantiated?
		if ac.Config.Labels == nil {
			ac.Config.Labels = make(map[string]string)
		}
		ac.Config.Labels["zodiacManifest"] = string(manifestJSON)
		id, err := client.CreateContainer(&ac.Config, ac.Name)
		if err != nil {
			log.Fatal("problem creating: ", err)
		}
		log.Infof("%s created as %s", ac.Name, id)
		if err := client.StartContainer(id, &dockerclient.HostConfig{}); err != nil {
			log.Fatal("problem starting: ", err)
		}
		log.Infof("started container '%s'", ac.Name)
	}

	return prettycli.PlainOutput{"TODO: something something something success."}, nil
}

func resolveImage(s string) (string, error) {
	image := s
	tag := "latest"
	if strings.ContainsRune(s, ':') {
		elements := strings.Split(s, ":")
		if len(elements) != 2 {
			log.Fatalf("can't find image and tag name from '%s'", s)
		}
		image = elements[0]
		tag = elements[1]
	}

	auth, err := RegistryClient.Hub.GetReadToken(image)
	if err != nil {
		return "", err
	}

	tags, err := RegistryClient.Repository.ListTags(image, auth)
	if err != nil {
		log.Fatalf("the image '%s' could not be found: %s", image, err.Error())
	}

	sha := tags[tag]
	if sha == "" {
		log.Fatalf("the tag '%s' couldn't be found for image '%s'", tag, image)
	}

	return sha, nil
}
