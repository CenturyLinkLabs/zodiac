package main

import (
	"testing"

	"github.com/CenturyLinkLabs/clitest"
	"github.com/stretchr/testify/assert"
)

var b clitest.BuildTester

func init() {
	b = clitest.NewBuild()
}

func TestNoArguments(t *testing.T) {
	r := b.Run(t)
	r.AssertSuccessful()
	assert.Equal(t, "No Arguments Passed", r.Stdout())
	assert.Empty(t, r.Stderr())
}

func TestArguments(t *testing.T) {
	r := b.Run(t, "-test")
	r.AssertSuccessful()
	assert.Equal(t, "You set the test flag", r.Stdout())
	assert.Empty(t, r.Stderr())
}

func TestEnvironmentVariables(t *testing.T) {
	r := b.RunWithOptions(t,
		clitest.RunOptions{
			Environment: map[string]string{
				"CLITEST_TEST_VAR": "testing123",
			},
		},
	)
	r.AssertSuccessful()
	assert.Equal(t, "CLITEST_TEST_VAR is testing123", r.Stdout())
	assert.Empty(t, r.Stderr())
}

func TestBadExit(t *testing.T) {
	r := b.Run(t, "-explode")
	r.AssertExitCode(19)
	assert.Equal(t, "I exploded", r.Stderr())
	assert.Empty(t, r.Stdout())
}
