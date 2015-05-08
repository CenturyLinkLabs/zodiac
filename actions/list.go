package actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/discovery"
)

func List(c discovery.Cluster) (prettycli.Output, error) {
	manis, err := manifestsFromCluster(c)
	if err != nil {
		return prettycli.PlainOutput{}, fmt.Errorf("unable to retrieve manifests from cluster: %s", err)
	}
	output := prettycli.ListOutput{
		Labels: []string{"#", "Time", "Names"},
	}

	for i, mani := range manis {
		names := make([]string, 0)
		for name := range mani.Services {
			names = append(names, name)
		}

		output.AddRow(map[string]string{
			"#":     strconv.Itoa(i + 1),
			"Time":  mani.Time,
			"Names": strings.Join(names, ", "),
		})
	}
	return output, nil
}
