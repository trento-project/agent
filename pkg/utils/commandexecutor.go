package utils

import (
	"context"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
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
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	go func() {
		<-ctx.Done()
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err != nil {
			log.Errorf("Error killing process group: %v", err)
		}
	}()

	return cmd.Output()
}
