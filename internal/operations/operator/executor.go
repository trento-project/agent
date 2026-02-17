package operator

import (
	"context"
	"errors"
	"log/slog"
)

type phaser interface {
	plan(ctx context.Context) (alreadyApplied bool, err error)
	commit(ctx context.Context) error
	rollback(ctx context.Context) error
	verify(ctx context.Context) error
	operationDiff(ctx context.Context) map[string]any
	after(ctx context.Context)
}

type Executor struct {
	currentPhase PhaseName
	phaser       phaser
	operationID  string
	logger       *slog.Logger
}

const (
	RUN     = "Executor.Run"
	BEGIN   = "BEGIN"
	SUCCESS = "SUCCESS"
	FAILURE = "FAILURE"
)

func NewExecutor(phaser phaser, operationID string, logger *slog.Logger) *Executor {
	if logger == nil {
		logger = slog.Default()
	}
	return &Executor{
		currentPhase: PLAN,
		phaser:       phaser,
		operationID:  operationID,
		logger:       logger,
	}
}

func (e *Executor) Run(ctx context.Context) *ExecutionReport {
	e.currentPhase = PLAN
	e.logger.Info(RUN, "phase", e.currentPhase, "event", BEGIN)
	alreadyApplied, err := e.phaser.plan(ctx)
	if err != nil {
		e.logger.Info(RUN, "phase", e.currentPhase, "event", FAILURE, "error", err)
		return executionReportWithError(err, e.currentPhase, e.operationID)
	}

	defer e.phaser.after(ctx)

	if alreadyApplied {
		diff := e.phaser.operationDiff(ctx)
		e.logger.Info(RUN, "phase", e.currentPhase, "event", SUCCESS, "diff", diff)
		return executionReportWithSuccess(diff, e.currentPhase, e.operationID)
	}
	e.logger.Info(RUN, "phase", e.currentPhase, "event", SUCCESS)

	e.currentPhase = COMMIT

	e.logger.Info(RUN, "phase", e.currentPhase, "event", BEGIN)
	err = e.phaser.commit(ctx)
	if err != nil {
		e.logger.Info(RUN, "phase", e.currentPhase, "event", FAILURE, "error", err)
		return e.handleRollback(ctx, err)
	}
	e.logger.Info(RUN, "phase", e.currentPhase, "event", SUCCESS)

	e.currentPhase = VERIFY
	e.logger.Info(RUN, "phase", e.currentPhase, "event", BEGIN)
	err = e.phaser.verify(ctx)
	if err != nil {
		e.logger.Info(RUN, "phase", e.currentPhase, "event", FAILURE, "error", err)
		return e.handleRollback(ctx, err)
	}

	diff := e.phaser.operationDiff(ctx)
	e.logger.Info(RUN, "phase", e.currentPhase, "event", SUCCESS, "diff", diff)

	return executionReportWithSuccess(diff, e.currentPhase, e.operationID)
}

func (e *Executor) handleRollback(ctx context.Context, err error) *ExecutionReport {
	e.logger.Info(RUN, "phase", ROLLBACK, "event", BEGIN)
	rollbackError := e.phaser.rollback(ctx)
	if rollbackError != nil {
		e.currentPhase = ROLLBACK
		e.logger.Info(RUN, "phase", e.currentPhase, "event", FAILURE, "error", rollbackError)
		return executionReportWithError(
			wrapRollbackError(err, rollbackError),
			e.currentPhase,
			e.operationID,
		)
	}
	e.logger.Info(RUN, "phase", ROLLBACK, "event", SUCCESS)
	return executionReportWithError(err, e.currentPhase, e.operationID)
}

func wrapRollbackError(phaseError error, rollbackError error) error {
	return errors.Join(rollbackError, phaseError)
}
