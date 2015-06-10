package composer

type Composer interface {
	Run(...string) error
}

type ExecComposer struct {
}

func NewExecComposer(dockerHost string) ExecComposer {
	// TODO: Implement. Don't have this do anything resource-intensive since it runs at init.
	return ExecComposer{}
}

func (c *ExecComposer) Run(args ...string) error {
	return nil
}
