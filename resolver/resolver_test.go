package resolver

import (
	"errors"
	"testing"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
	"github.com/stretchr/testify/assert"
)

type HardcodedImageResolver struct {
	CalledImage   string
	ListTagsError error
	tags          registry.TagMap
}

func (c *HardcodedImageResolver) ListTags(image string) (registry.TagMap, error) {
	c.CalledImage = image
	if c.ListTagsError != nil {
		return make(registry.TagMap), c.ListTagsError
	}
	return c.tags, nil
}

func TestResolveImage_SuccessfulDefaultTag(t *testing.T) {
	r := HardcodedImageResolver{tags: registry.TagMap{"latest": "testsha"}}
	DefaultImageResolver = &r

	sha, err := ResolveImage("example/test")
	assert.Equal(t, "example/test:latest", r.CalledImage)
	assert.NoError(t, err)
	assert.Equal(t, "testsha", sha)
}

func TestResolveImage_SuccessfulExplicitTag(t *testing.T) {
	r := HardcodedImageResolver{tags: registry.TagMap{"0.1": "testsha"}}
	DefaultImageResolver = &r

	sha, err := ResolveImage("example/test:0.1")
	assert.Equal(t, "example/test:0.1", r.CalledImage)
	assert.NoError(t, err)
	assert.Equal(t, "testsha", sha)
}

func TestResolveImage_ErroredCrazyName(t *testing.T) {
	r := HardcodedImageResolver{}
	DefaultImageResolver = &r

	sha, err := ResolveImage("example/test:0.1:0.2")
	assert.Empty(t, r.CalledImage)
	assert.Empty(t, sha)
	assert.EqualError(t, err, "can't find image and tag name from 'example/test:0.1:0.2'")
}

func TestResolveImage_ErroredListTags(t *testing.T) {
	r := HardcodedImageResolver{ListTagsError: errors.New("test error")}
	DefaultImageResolver = &r

	sha, err := ResolveImage("example/test")
	assert.Empty(t, sha)
	assert.EqualError(t, err, "the image 'example/test' could not be found: test error")
}

func TestResolveImage_ErroredTagNotFound(t *testing.T) {
	r := HardcodedImageResolver{tags: registry.TagMap{}}
	DefaultImageResolver = &r

	sha, err := ResolveImage("example/test")
	assert.Empty(t, sha)
	assert.EqualError(t, err, "the tag 'latest' couldn't be found for image 'example/test'")
}
