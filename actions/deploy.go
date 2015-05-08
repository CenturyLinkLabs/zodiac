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
	"time"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/discovery"
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

var RegistryClient = registry.NewClient()

const SavedDeploymentCount = 5

type AttemptedContainer struct {
	Name          string
	CreateOptions []byte
	Config        dockerclient.ContainerConfig
}

func Deploy(c discovery.Cluster) (prettycli.Output, error) {
	attemptedContainers := make([]*AttemptedContainer, 0)

	//////////////////
	// WAIT FOR COMPOSE TO DEPLOY
	//////////////////
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bs, _ := ioutil.ReadAll(r.Body)
		log.Infof("Got: %s %s, Body: %s", r.Method, r.URL.String(), string(bs))

		// TODO Audit that URL and meth are specific enough
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
		} else if r.Method == "GET" && r.URL.Path == "/v1.15/containers/json" {
			// TODO you've introduced a bug where it'll try and create containers
			// overtop of other identically-named containers because it never knows
			// they are running. Good job.
			fmt.Fprintf(w, `[]`)
		} else {
			log.Fatalf("got unhandled request '%s'", r.URL.String())
		}
	})

	fmt.Println("So uhhhh, press any key when you're deployed... Or Ctrl-C if you want to abort...")
	laddr, _ := net.ResolveTCPAddr("tcp", "localhost:31981")
	listener, _ := net.ListenTCP("tcp", laddr)
	go http.Serve(listener, handler)
	defer listener.Close()
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	//////////////////
	// FETCHING OLD DEPLOYMENT
	//////////////////

	oldManifests, err := manifestsFromCluster(c)
	if err != nil {
		log.Fatal("unable to retrieve manifests from cluster: ", err)
	}

	if err := removeContainers(c); err != nil {
		log.Fatal("unable to remove containers: ", err)
	}

	//////////////////
	// YOU NOW HAVE THE CONTAINERS SAVED AND COMPOSE IS HAPPY
	//////////////////

	fmt.Println("Starting containers...")
	for _, ac := range attemptedContainers {
		var cc dockerclient.ContainerConfig
		if err := json.Unmarshal(ac.CreateOptions, &cc); err != nil {
			log.Fatalf("error unmarshalling request JSON for '%s': %s", ac.Name, err.Error())
		}
		ac.Config = cc
	}
	currentManifest := NewDeploymentManifest()
	currentManifest.Time = time.Now().String()
	for _, ac := range attemptedContainers {
		sha, err := resolveImage(ac.Config.Image)
		if err != nil {
			log.Fatalf("the image '%s' couldn't be found: %s", ac.Config.Image, err.Error())
		}
		currentManifest.Services[ac.Name] = sha
	}
	oldDeploymentCount := len(oldManifests)
	if oldDeploymentCount >= SavedDeploymentCount {
		oldManifests = oldManifests[(oldDeploymentCount - SavedDeploymentCount + 1):]
	}
	finalManifest := append(oldManifests, currentManifest)
	finalManifestJSON, err := json.Marshal(finalManifest)
	if err != nil {
		log.Fatal("there was a problem building the deployment manifest: ", err)
	}

	for _, ep := range c.Endpoints() {
		client, err := dockerclient.NewDockerClient(ep.URL, nil)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("creating containers on endpoints '%s'", ep.URL)

		for _, ac := range attemptedContainers {
			// TODO Why? Change the way the Config is instantiated?
			if ac.Config.Labels == nil {
				ac.Config.Labels = make(map[string]string)
			}
			ac.Config.Labels["zodiacManifest"] = string(finalManifestJSON)
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
