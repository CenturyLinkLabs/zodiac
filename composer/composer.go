package composer

import (
	"bytes"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

type Composer interface {
	Run() error
}

type ExecComposer struct {
}

func NewExecComposer(dockerHost string) *ExecComposer {
	return &ExecComposer{}
}

func (c *ExecComposer) DrainRequests(args []string) error {
	return nil
}

func (c *ExecComposer) Run() error {
	// TODO: implement --file functionality
	composeArgs := []string{"up", "-d"}
	cmd := exec.Command("docker-compose", composeArgs...)
	// TODO: this port must match the proxy port, see related TODO in cluster/proxy.go
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
