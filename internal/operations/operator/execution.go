package operator

import (
	"fmt"
)

type ExecutionError struct {
	ErrorPhase PhaseName
	Message    string
}

func (e ExecutionError) Error() string {
	return fmt.Sprintf(
		"error during operator exeuction in phase: %s, reason: %s",
		e.ErrorPhase,
		e.Message,
	)
}

type ExecutionSuccess struct {
	Diff      map[string]any
	LastPhase PhaseName
}

type ExecutionReport struct {
	OperationID string
	Success     *ExecutionSuccess
	Error       *ExecutionError
}

func executionReportWithError(err error, phase PhaseName, operationID string) *ExecutionReport {
	return &ExecutionReport{
		OperationID: operationID,
		Error: &ExecutionError{
			Message:    err.Error(),
			ErrorPhase: phase,
		},
	}
}

func executionReportWithSuccess(diff map[string]any, phase PhaseName, operationID string) *ExecutionReport {
	return &ExecutionReport{
		OperationID: operationID,
		Success: &ExecutionSuccess{
			Diff:      diff,
			LastPhase: phase,
		},
	}
}
