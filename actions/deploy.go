package actions

import (
	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/discovery"
)

func Deploy(c discovery.Cluster) (prettycli.Output, error) {
	return prettycli.PlainOutput{"whatevs"}, nil
}
