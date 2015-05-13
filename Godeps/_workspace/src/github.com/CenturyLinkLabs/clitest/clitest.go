package clitest

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

const tmpPath = "/tmp/clitest"

type BuildTester struct {
	binaryPath string
}

type RunOptions struct {
	Arguments   []string
	Environment map[string]string
}

type TestRun struct {
	test   *testing.T
	result error
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func NewBuild() BuildTester {
	cmd := exec.Command("go", "build", "-o", tmpPath, ".")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("couldn't build your program:\n%s\n\n%s\n", err, stderr.String()))
	}

	return BuildTester{
		binaryPath: tmpPath,
	}
}

func (t *BuildTester) Run(test *testing.T, args ...string) TestRun {
	return t.RunWithOptions(test,
		RunOptions{
			Arguments:   args,
			Environment: make(map[string]string),
		},
	)
}

func (t *BuildTester) RunWithOptions(test *testing.T, options RunOptions) TestRun {
	r := TestRun{
		test:   test,
		stdout: bytes.NewBuffer([]byte{}),
		stderr: bytes.NewBuffer([]byte{}),
	}

	cmd := exec.Command(t.binaryPath, options.Arguments...)
	for name, value := range options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", name, value))
	}
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	r.result = cmd.Run()

	return r
}

func (t TestRun) AssertSuccessful() bool {
	if t.result != nil {
		t.test.Errorf("expected a successful run result, got: %s", t.result)
		return false
	}

	return true
}

func (t TestRun) AssertExitCode(code int) bool {
	expected := fmt.Sprintf("exit status %d", code)
	if t.result != nil && t.result.Error() != expected {
		t.test.Errorf("\nexpected:\n%s\ngot:\n%s", expected, t.result.Error())
	}

	return true
}

func (t TestRun) Stdout() string {
	return t.stdout.String()
}

func (t TestRun) Stderr() string {
	return t.stderr.String()
}
