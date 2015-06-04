package composer

import (
	"bytes"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

type Composer interface {
	Run([]string) error
}

type ExecComposer struct {
}

func NewExecComposer(dockerHost string) *ExecComposer {
	// TODO: Implement. Don't have this do anything resource-intensive since it runs at init.
	return &ExecComposer{}
}

func (c *ExecComposer) DrainRequests(args []string) error {
	return nil
}

func (c *ExecComposer) Run(args []string) error {
	combinedArgs := append([]string{"up", "-d"}, args...)
	cmd := exec.Command("docker-compose", combinedArgs...)
	// TODO: this port must match the proxy port
	cmd.Env = []string{"DOCKER_HOST=localhost:3000"}
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
