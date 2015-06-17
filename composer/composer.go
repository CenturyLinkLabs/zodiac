package composer

import (
	"bytes"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

type Composer interface {
	Run(map[string]string) error
}

type ExecComposer struct {
}

func NewExecComposer(dockerHost string) *ExecComposer {
	return &ExecComposer{}
}

func (c *ExecComposer) DrainRequests(args []string) error {
	return nil
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
	cmd.Env = []string{"DOCKER_HOST=localhost:61908"}
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()

	log.Info("docker-compose stdout: ", out.String())
	if err != nil {
		log.Info("docker-compose stderr: ", errOut.String())
		log.Fatal(err)
	}

	return err
}
