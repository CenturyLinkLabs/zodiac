package main

import (
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/CenturyLinkLabs/clitest"
	"github.com/CenturyLinkLabs/zodiac/fakeengine"
	"github.com/stretchr/testify/assert"
)

var b *clitest.BuildTester

const TLSFlag = "--tls=false"

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

	r := b.Run(t, TLSFlag, endpointFlag, "verify")
	r.AssertSuccessful()
	assert.Contains(t, r.Stdout(), "Successfully verified endpoint:")
	assert.Empty(t, r.Stderr())
}

func TestVerify_NoEndpoint(t *testing.T) {
	setup(t)

	r := b.Run(t, "verify")
	r.AssertExitCode(1)
	assert.Contains(t, r.Stderr(), "specify a Docker endpoint")
	assert.Empty(t, r.Stdout())
}

func TestVerify_EndpointEnvVar(t *testing.T) {
	setup(t)
	s, endpointFlag := newFakeServerAndFlag()
	defer s.Close()

	parts := strings.Split(endpointFlag, "=")

	os.Setenv("DOCKER_HOST", parts[1])
	defer os.Unsetenv("DOCKER_HOST")

	r := b.Run(t, TLSFlag, "verify")
	r.AssertSuccessful()
	assert.Contains(t, r.Stdout(), "Successfully verified endpoint:")
	assert.Empty(t, r.Stderr())
}

func TestDeploy_Successful(t *testing.T) {
	setup(t)
	s, endpointFlag := newFakeServerAndFlag()
	defer s.Close()

	r := b.Run(t, TLSFlag, endpointFlag, "deploy", "-f", "fixtures/webapp.yml")
	fmt.Println(r.Stderr())
	fmt.Println(r.Stdout())
	r.AssertSuccessful()
	assert.Contains(t, r.Stdout(), "Successfully deployed 2 container(s)")
	assert.Empty(t, r.Stderr())
}

func TestTeardown_Prompt(t *testing.T) {
	setup(t)
	s, endpointFlag := newFakeServerAndFlag()
	defer s.Close()

	r := b.Run(t, TLSFlag, endpointFlag, "teardown", "-f", "fixtures/webapp.yml")
	fmt.Println(r.Stderr())
	fmt.Println(r.Stdout())
	r.AssertExitCode(1)
	assert.Contains(t, r.Stdout(), "Are you sure you want to remove the deployment and all history? (y/N)")
}

func TestTeardown_Success(t *testing.T) {
	setup(t)
	s, endpointFlag := newFakeServerAndFlag()
	defer s.Close()

	r := b.Run(t, TLSFlag, endpointFlag, "teardown", "-f", "fixtures/webapp.yml", "-c", "yes")
	fmt.Println(r.Stderr())
	fmt.Println(r.Stdout())
	r.AssertSuccessful()
	assert.Contains(t, r.Stdout(), "Successfully removed 2 services and all deployment history")
	assert.Empty(t, r.Stderr())
}
