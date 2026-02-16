package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
)

const (
	SapInstanceStopOperatorName = "sapinstancestop"
)

type sapInstanceStopDiffOutput struct {
	Stopped bool `json:"stopped"`
}

type SAPInstanceStopOption Option[SAPInstanceStop]

type SAPInstanceStop struct {
	baseOperator
	parsedArguments     *sapStateChangeArguments
	sapControlConnector sapcontrolapi.WebService
	interval            time.Duration
}

func WithCustomStopSapcontrol(sapControlConnector sapcontrolapi.WebService) SAPInstanceStopOption {
	return func(o *SAPInstanceStop) {
		o.sapControlConnector = sapControlConnector
	}
}

func WithCustomStopInterval(interval time.Duration) SAPInstanceStopOption {
	return func(o *SAPInstanceStop) {
		o.interval = interval
	}
}

// NewSAPInstanceStop operator stops a SAP instance.
//
// Arguments:
//  instance_number (required): String with the instance number of the instance to stop
//  timeout: Timeout in seconds to wait until the instance is stopped
//
// # Execution Phases
//
// - PLAN:
//   The operator gets the instance current processes and stores the state.
//   The operation is skipped if the SAP instances is already stopped.
//
// - COMMIT:
//   It stops the SAP instance using the sapcontrol Stop command.
//
// - VERIFY:
//   Verify if the SAP instance is stopped.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the instance is started back again.

func NewSAPInstanceStop(
	arguments Arguments,
	operationID string,
	options Options[SAPInstanceStop],
) *Executor {
	sapInstanceStop := &SAPInstanceStop{
		baseOperator: newBaseOperator(
			SapInstanceStopOperatorName, operationID, arguments, options.BaseOperatorOptions...,
		),
		interval: defaultSapInstanceStateInterval,
	}

	for _, opt := range options.OperatorOptions {
		opt(sapInstanceStop)
	}

	return &Executor{
		phaser:      sapInstanceStop,
		operationID: operationID,
		logger:      sapInstanceStop.logger,
	}
}

func (s *SAPInstanceStop) plan(ctx context.Context) (bool, error) {
	opArguments, err := parseSAPStateChangeArguments(s.arguments)
	if err != nil {
		return false, err
	}
	s.parsedArguments = opArguments

	// Use custom sapControlConnector or create a new one based on the instance_number argument
	if s.sapControlConnector == nil {
		s.sapControlConnector = sapcontrolapi.NewWebServiceUnix(s.parsedArguments.instNumber)
	}

	stopped, err := allProcessesInState(ctx, s.sapControlConnector, sapcontrolapi.STATECOLOR_GRAY)
	if err != nil {
		return false, fmt.Errorf("error checking processes state: %w", err)
	}

	s.resources[beforeDiffField] = stopped

	if stopped {
		s.logger.Info("instance already stopped, skipping operation")
		s.resources[afterDiffField] = stopped
		return true, nil
	}

	return false, nil
}

func (s *SAPInstanceStop) commit(ctx context.Context) error {
	request := new(sapcontrolapi.Stop)
	_, err := s.sapControlConnector.StopContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error stopping instance: %w", err)
	}

	return nil
}

func (s *SAPInstanceStop) verify(ctx context.Context) error {
	err := waitUntilSapInstanceState(
		ctx,
		s.sapControlConnector,
		sapcontrolapi.STATECOLOR_GRAY,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("verify instance stopped failed: %w", err)
	}

	s.resources[afterDiffField] = true
	return nil
}

func (s *SAPInstanceStop) rollback(ctx context.Context) error {
	request := new(sapcontrolapi.Start)
	_, err := s.sapControlConnector.StartContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting instance: %w", err)
	}

	err = waitUntilSapInstanceState(
		ctx,
		s.sapControlConnector,
		sapcontrolapi.STATECOLOR_GREEN,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("rollback to started failed: %w", err)
	}

	return nil
}

//	operationDiff needs to be refactored, ignoring duplication issues for now
//
// nolint: dupl
func (s *SAPInstanceStop) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeStopped, ok := s.resources[beforeDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid beforeStopped value: cannot parse '%s' to bool",
			s.resources[beforeDiffField]))
	}

	afterStopped, ok := s.resources[afterDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid afterStopped value: cannot parse '%s' to bool",
			s.resources[afterDiffField]))
	}

	beforeDiffOutput := sapInstanceStopDiffOutput{
		Stopped: beforeStopped,
	}
	before, err := json.Marshal(beforeDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling before diff output: %v", err))
	}
	diff["before"] = string(before)

	afterDiffOutput := sapInstanceStopDiffOutput{
		Stopped: afterStopped,
	}
	after, err := json.Marshal(afterDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling after diff output: %v", err))
	}
	diff["after"] = string(after)

	return diff
}
