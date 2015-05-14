package resolver

import (
	"fmt"
	"strings"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
)

var DefaultImageResolver Resolver

type Resolver interface {
	ListTags(string) (registry.TagMap, error)
}

type DockerRegClientResolver struct{}

func (c DockerRegClientResolver) ListTags(image string) (registry.TagMap, error) {
	client := registry.NewClient()
	auth, err := client.Hub.GetReadToken(image)
	if err != nil {
		return make(registry.TagMap), err
	}

	tags, err := client.Repository.ListTags(image, auth)
	if err != nil {
		return make(registry.TagMap), err
	}

	return tags, nil
}

func ResolveImage(definition string) (string, error) {
	image := definition
	tag := "latest"
	if strings.ContainsRune(definition, ':') {
		elements := strings.Split(definition, ":")
		if len(elements) != 2 {
			return "", fmt.Errorf("can't find image and tag name from '%s'", definition)
		}
		image = elements[0]
		tag = elements[1]
	}

	tags, err := DefaultImageResolver.ListTags(fmt.Sprintf("%s:%s", image, tag))
	if err != nil {
		return "", fmt.Errorf("the image '%s' could not be found: %s", image, err.Error())
	}

	sha := tags[tag]
	if sha == "" {
		return "", fmt.Errorf("the tag '%s' couldn't be found for image '%s'", tag, image)
	}

	return sha, nil
}
