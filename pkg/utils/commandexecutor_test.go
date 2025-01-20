package utils_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	commandexecutor "github.com/trento-project/agent/pkg/utils"
)

func TestExec(t *testing.T) {
	executor := commandexecutor.Executor{}

	result, err := executor.Exec("echo", "trento is not trieste")
	if err != nil {
		t.Errorf("Error executing command: %v", err)
	}

	assert.Equal(t, "trento is not trieste\n", string(result))
}

func TestExecWithError(t *testing.T) {
	executor := commandexecutor.Executor{}

	_, err := executor.Exec("nonexistentcommand")
	if err == nil {
		t.Errorf("Expected error executing command")
	}
}

func TestExecContext(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx := context.Background()

	result, err := executor.ExecContext(ctx, "echo", "trento is not trieste")
	if err != nil {
		t.Errorf("Error executing command: %v", err)
	}

	assert.Equal(t, "trento is not trieste\n", string(result))
}

func TestExecContextWithError(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx := context.Background()

	_, err := executor.ExecContext(ctx, "nonexistentcommand")
	if err == nil {
		t.Errorf("Expected error executing command")
	}
}

func TestExecContextWithCancel(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	_, err := executor.ExecContext(ctx, "echo", "trento is not trieste")
	if err == nil {
		t.Errorf("Expected error executing command")
	}
}

func TestExecContextWithCancelLongRunning(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx, cancel := context.WithCancel(context.Background())
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelTimeout()

	go func() {
		cancel()
	}()

	_, err := executor.ExecContext(ctx, "sleep", "3s")

	switch {
	case timeoutCtx.Err() != nil:
		t.Errorf("Expected process to be killed before timeout")
	case err == nil:
		t.Errorf("Expected error when cancelling")
	}

}
