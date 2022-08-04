package gatherers

import (
	"os/exec"
)

//go:generate mockery --name=CommandExecutor

type CommandExecutor interface {
	Exec(name string, arg ...string) ([]byte, error)
}

type Executor struct{}

func (e Executor) Exec(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	return cmd.Output()
}
