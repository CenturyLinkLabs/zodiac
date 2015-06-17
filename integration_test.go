package main

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CenturyLinkLabs/clitest"
	"github.com/CenturyLinkLabs/zodiac/fakeengine"
	"github.com/stretchr/testify/assert"
)

var b *clitest.BuildTester

func setup(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	if b == nil {
		tmp := clitest.NewBuild()
		b = &tmp
	}
}

func newFakeServerAndFlag() (*httptest.Server, string) {
	s := fakeengine.NewServer()
	elements := strings.Split(s.URL, ":")
	if len(elements) != 3 {
		panic("there was a problem with the test server!")
	}
	flag := fmt.Sprintf(`--endpoint=tcp://localhost:%s`, elements[2])

	return s, flag
}

func TestHelp_Successful(t *testing.T) {
	setup(t)
	r := b.Run(t)
	r.AssertSuccessful()
	assert.Contains(t, r.Stdout(), "Simple Docker deployment")
	assert.Empty(t, r.Stderr())
}

func TestVerify_Successful(t *testing.T) {
	setup(t)
	s, endpointFlag := newFakeServerAndFlag()
	defer s.Close()

	r := b.Run(t, endpointFlag, "verify")
	r.AssertSuccessful()
	assert.Contains(t, r.Stdout(), "Successfully verified endpoint:")
	assert.Empty(t, r.Stderr())
}

func TestVerify_NoCluster(t *testing.T) {
	setup(t)

	r := b.Run(t, "verify")
	r.AssertExitCode(1)
	assert.Contains(t, r.Stderr(), "specify a Docker endpoint")
	assert.Empty(t, r.Stdout())
}

func TestDeploy_Successful(t *testing.T) {
	setup(t)
	s, endpointFlag := newFakeServerAndFlag()
	defer s.Close()

	r := b.Run(t, endpointFlag, "deploy", "-f", "fixtures/webapp.yml")
	fmt.Println(r.Stderr())
	fmt.Println(r.Stdout())
	r.AssertSuccessful()
	assert.Contains(t, r.Stdout(), "Successfully deployed 2 container(s)")
	assert.Empty(t, r.Stderr())
}
