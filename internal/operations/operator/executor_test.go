package operator_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	operator "github.com/trento-project/agent/internal/operations/operator"
)

func TestExecutorHappyFlow(t *testing.T) {
	executionContext := context.Background()
	phaser := operator.NewMockphaser(t)
	emptyDiff := make(map[string]any)

	planCall := phaser.On("plan", executionContext).
		Return(false, nil)

	commitCall := phaser.On("commit", executionContext).
		Return(nil).
		NotBefore(planCall)

	verifyCall := phaser.On("verify", executionContext).
		Return(nil).
		NotBefore(commitCall)

	operationDiffCall := phaser.On("operationDiff", executionContext).
		Return(emptyDiff).
		NotBefore(verifyCall)

	phaser.On("after", executionContext).
		Return().
		Once().
		NotBefore(operationDiffCall)

	executor := operator.NewExecutor(phaser, "operation-id", slog.Default())

	report := executor.Run(executionContext)

	assert.Equal(t, "operation-id", report.OperationID)
	assert.Equal(t, operator.VERIFY, report.Success.LastPhase)
	assert.Equal(t, emptyDiff, report.Success.Diff)
	assert.Nil(t, report.Error)
}

func TestExecutorPlanError(t *testing.T) {
	executionContext := context.Background()
	phaser := operator.NewMockphaser(t)
	planError := errors.New("error during plan phase")

	phaser.On("plan", executionContext).
		Return(false, planError)

	phaser.AssertNotCalled(t, "after", executionContext)

	executor := operator.NewExecutor(phaser, "operation-id", slog.Default())

	report := executor.Run(executionContext)

	assert.Equal(t, planError.Error(), report.Error.Message)
	assert.Equal(t, operator.PLAN, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}

func TestExecutorPlanAlreadyApplied(t *testing.T) {
	executionContext := context.Background()
	phaser := operator.NewMockphaser(t)
	emptyDiff := make(map[string]any)

	planCall := phaser.On("plan", executionContext).
		Return(true, nil)

	operationDiffCall := phaser.On("operationDiff", executionContext).
		Return(emptyDiff).
		NotBefore(planCall)

	phaser.On("after", executionContext).
		Return().
		Once().
		NotBefore(operationDiffCall)

	executor := operator.NewExecutor(phaser, "operation-id", slog.Default())

	report := executor.Run(executionContext)

	assert.Equal(t, "operation-id", report.OperationID)
	assert.Equal(t, operator.PLAN, report.Success.LastPhase)
	assert.Equal(t, emptyDiff, report.Success.Diff)
	assert.Nil(t, report.Error)
}

func TestExecutorCommitErrorWithSuccessfulRollback(t *testing.T) {
	executionContext := context.Background()
	phaser := operator.NewMockphaser(t)
	commitError := errors.New("error during error phase")

	planCall := phaser.On("plan", executionContext).
		Return(false, nil)

	commitCall := phaser.On("commit", executionContext).
		Return(commitError).
		NotBefore(planCall)

	rollbackCall := phaser.On("rollback", executionContext).
		Return(nil).
		NotBefore(commitCall)

	phaser.On("after", executionContext).
		Return().
		Once().
		NotBefore(rollbackCall)

	executor := operator.NewExecutor(phaser, "operation-id", slog.Default())

	report := executor.Run(executionContext)

	assert.Equal(t, commitError.Error(), report.Error.Message)
	assert.Equal(t, operator.COMMIT, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}

func TestExecutorCommitErrorWithFailedRollback(t *testing.T) {
	executionContext := context.Background()
	phaser := operator.NewMockphaser(t)
	commitError := errors.New("error during error phase")
	rollbackError := errors.New("error during rollback phase")

	planCall := phaser.On("plan", executionContext).
		Return(false, nil)

	commitCall := phaser.On("commit", executionContext).
		Return(commitError).
		NotBefore(planCall)

	rollbackCall := phaser.On("rollback", executionContext).
		Return(rollbackError).
		NotBefore(commitCall)

	phaser.On("after", executionContext).
		Return().
		Once().
		NotBefore(rollbackCall)

	executor := operator.NewExecutor(phaser, "operation-id", slog.Default())

	report := executor.Run(executionContext)

	assert.Equal(t, errors.Join(rollbackError, commitError).Error(), report.Error.Message)
	assert.Equal(t, operator.ROLLBACK, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}

func TestExecutorVerifyErrorWithSuccessfulRollback(t *testing.T) {
	executionContext := context.Background()
	phaser := operator.NewMockphaser(t)
	verifyError := errors.New("error during verify phase")

	planCall := phaser.On("plan", executionContext).
		Return(false, nil)

	commitCall := phaser.On("commit", executionContext).
		Return(nil).
		NotBefore(planCall)

	verifyCall := phaser.On("verify", executionContext).
		Return(verifyError).
		NotBefore(commitCall)

	rollbackCall := phaser.On("rollback", executionContext).
		Return(nil).
		NotBefore(verifyCall)

	phaser.On("after", executionContext).
		Return().
		Once().
		NotBefore(rollbackCall)

	executor := operator.NewExecutor(phaser, "operation-id", slog.Default())

	report := executor.Run(executionContext)

	assert.Equal(t, verifyError.Error(), report.Error.Message)
	assert.Equal(t, operator.VERIFY, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}

func TestExecutorVerifyErrorWithFailedRollback(t *testing.T) {
	executionContext := context.Background()
	phaser := operator.NewMockphaser(t)
	verifyError := errors.New("error during verify phase")
	rollbackError := errors.New("error during rollback phase")

	planCall := phaser.On("plan", executionContext).
		Return(false, nil)

	commitCall := phaser.On("commit", executionContext).
		Return(nil).
		NotBefore(planCall)

	verifyCall := phaser.On("verify", executionContext).
		Return(verifyError).
		NotBefore(commitCall)

	rollbackCall := phaser.On("rollback", executionContext).
		Return(rollbackError).
		NotBefore(verifyCall)

	phaser.On("after", executionContext).
		Return().
		Once().
		NotBefore(rollbackCall)

	executor := operator.NewExecutor(phaser, "operation-id", slog.Default())

	report := executor.Run(executionContext)

	assert.Equal(t, errors.Join(rollbackError, verifyError).Error(), report.Error.Message)
	assert.Equal(t, operator.ROLLBACK, report.Error.ErrorPhase)
	assert.Nil(t, report.Success)
}
