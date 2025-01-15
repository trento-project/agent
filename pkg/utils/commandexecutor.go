package utils

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
)

//go:generate mockery --name=CommandExecutor

type CommandExecutor interface {
	Exec(name string, arg ...string) ([]byte, error)
	ExecContext(ctx context.Context, name string, arg ...string) ([]byte, error)
}

type Executor struct{}

func (e Executor) Exec(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)

	return cmd.Output()
}

func (e Executor) ExecContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		<-ctx.Done()
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return fmt.Errorf("error killing process group: %w", err)
	}
	return cmd.Output()
}
