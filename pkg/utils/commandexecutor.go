package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type CommandExecutor interface {
	Exec(name string, arg ...string) ([]byte, error)
	ExecContext(ctx context.Context, name string, arg ...string) ([]byte, error)
}

type Executor struct{}

func (e Executor) Exec(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "LC_ALL=C")

	return cmd.Output()
}

func (e Executor) ExecContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err != nil {
			return fmt.Errorf("error killing process group: %w", err)
		}
		return nil
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "LC_ALL=C")

	return cmd.Output()
}
