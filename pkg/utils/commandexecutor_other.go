// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

//go:build windows

// This file usually won't be compiled, as we solely support Linux builds.
// However, developers may want to build the agent on Windows for testing purposes,
// and this file is needed to avoid build errors when building on Windows.

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
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	cmd.Cancel = func() error {
		err := cmd.Process.Kill()
		if err != nil {
			return fmt.Errorf("error killing process: %w", err)
		}

		return nil
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "LC_ALL=C")

	return cmd
}
