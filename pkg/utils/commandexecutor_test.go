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

func TestOutput(t *testing.T) {
	executor := commandexecutor.Executor{}

	result, err := executor.Output("echo", "trento is not trieste")

	assert.NoError(t, err)
	assert.Equal(t, "trento is not trieste\n", string(result))
}

func TestOutputWithError(t *testing.T) {
	executor := commandexecutor.Executor{}

	_, err := executor.Output("nonexistentcommand")

	assert.Error(t, err)
}

func TestOutputContext(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx := context.Background()

	result, err := executor.OutputContext(ctx, "echo", "trento is not trieste")

	assert.NoError(t, err)
	assert.Equal(t, "trento is not trieste\n", string(result))
}

func TestOutputContextOnlyStdout(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx := context.Background()

	result, err := executor.OutputContext(ctx, "sh", "-c", `echo "This is stdout"; echo "This is stderr" >&2`)

	assert.NoError(t, err)
	assert.Equal(t, "This is stdout\n", string(result))
}

func TestCombinedOutputContext(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx := context.Background()

	result, err := executor.CombinedOutputContext(ctx, "sh", "-c", `echo "This is stdout"; echo "This is stderr" >&2`)

	assert.NoError(t, err)
	assert.Equal(t, "This is stdout\nThis is stderr\n", string(result))
}

func TestOutputContextWithError(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx := context.Background()

	_, err := executor.OutputContext(ctx, "nonexistentcommand")

	assert.Error(t, err)
}

func TestOutputContextWithCancel(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	_, err := executor.OutputContext(ctx, "echo", "trento is not trieste")
	assert.Error(t, err)
}

func TestOutputContextWithCancelLongRunning(t *testing.T) {
	executor := commandexecutor.Executor{}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	defer cancelTimeout()

	_, err := executor.OutputContext(ctx, "sleep", "3s")
	assert.Error(t, err)
	assert.NoError(t, timeoutCtx.Err())
}

func TestOutputWithLC_ALL(t *testing.T) {
	executor := commandexecutor.Executor{}

	result, err := executor.Output("sh", "-c", "echo $LC_ALL")

	assert.NoError(t, err)
	assert.Equal(t, "C\n", string(result))
}

func TestOutputContextWithLC_ALL(t *testing.T) {
	executor := commandexecutor.Executor{}

	result, err := executor.OutputContext(context.Background(), "sh", "-c", "echo $LC_ALL")

	assert.NoError(t, err)
	assert.Equal(t, "C\n", string(result))
}
