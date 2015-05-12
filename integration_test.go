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
	if testing.Short() {
		t.SkipNow()
	}

	r := b.Run(t)
	r.AssertSuccessful()
	assert.Contains(t, r.Stdout(), "Simple Docker deployment")
	assert.Empty(t, r.Stderr())
}
