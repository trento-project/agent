package utils_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	commandexecutor "github.com/trento-project/agent/pkg/utils"
)

type CommandExecutorTestSuite struct {
	suite.Suite
}

func TestCommandExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(CommandExecutorTestSuite))
}

func TestExec(t *testing.T) {
	executor := commandexecutor.Executor{}

	result, err := executor.Exec("echo", "trento is not trieste")

	assert.NoError(t, err)
	assert.Equal(t, "trento is not trieste\n", string(result))
}

func TestExecWithError(t *testing.T) {
	executor := commandexecutor.Executor{}

	_, err := executor.Exec("nonexistentcommand")

	assert.Error(t, err)
}

func TestExecContext(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx := context.Background()

	result, err := executor.ExecContext(ctx, "echo", "trento is not trieste")

	assert.NoError(t, err)
	assert.Equal(t, "trento is not trieste\n", string(result))
}

func TestExecContextWithError(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx := context.Background()

	_, err := executor.ExecContext(ctx, "nonexistentcommand")

	assert.Error(t, err)
}

func TestExecContextWithCancel(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	_, err := executor.ExecContext(ctx, "echo", "trento is not trieste")
	assert.Error(t, err)
}

func TestExecContextWithCancelLongRunning(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	defer cancelTimeout()

	_, err := executor.ExecContext(ctx, "sleep", "3s")
	assert.Error(t, err)
	assert.NoError(t, timeoutCtx.Err())
}
