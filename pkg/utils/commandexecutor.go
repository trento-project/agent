package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type CommandExecutor interface {
	Output(name string, arg ...string) ([]byte, error)
	OutputContext(ctx context.Context, name string, arg ...string) ([]byte, error)
	CombinedOutputContext(ctx context.Context, name string, arg ...string) ([]byte, error)
}

type Executor struct{}

func (e Executor) Output(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "LC_ALL=C")

	return cmd.Output()
}

func (e Executor) OutputContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	cmd := commandContext(ctx, name, arg...)
	return cmd.Output()
}

func (e Executor) CombinedOutputContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	cmd := commandContext(ctx, name, arg...)
	return cmd.CombinedOutput()
}

func commandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
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
	return cmd
}
