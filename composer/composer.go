package composer

import (
	"bytes"
	"fmt"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

type Composer interface {
	Run(map[string]string) error
}

type ExecComposer struct {
	dockerHost string
}

func NewExecComposer(dockerHost string) *ExecComposer {
	return &ExecComposer{dockerHost: dockerHost}
}

func (c *ExecComposer) Run(flags map[string]string) error {
	composeArgs := []string{"up", "-d"}
	for key, value := range flags {
		if key == "name" {
			composeArgs = append([]string{"-p", value}, composeArgs...)
		}
		if key == "file" {
			composeArgs = append([]string{"-f", value}, composeArgs...)
		}
	}
	cmd := exec.Command("docker-compose", composeArgs...)
	cmd.Env = []string{fmt.Sprintf("DOCKER_HOST=%s", c.dockerHost)}
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()

	log.Info("docker-compose stdout: ", out.String())
	if err != nil {
		log.Info("docker-compose stderr: ", errOut.String())
	}

	return err
}
