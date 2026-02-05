package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
)

const (
	SapSystemStartOperatorName        = "sapsystemstart"
	defaultSapSystemStateTimeout      = 5 * time.Minute
	defaultSapSystemStateInterval     = 10 * time.Second
	defaultSapSystemStateInstanceType = sapcontrolapi.StartStopOptionSAPControlALLINSTANCES

	instanceTypeALL    = "all"
	instanceTypeABAP   = "abap"
	instanceTypeJ2EE   = "j2ee"
	instanceTypeSCS    = "scs"
	instanceTypeENQREP = "enqrep"
)

type sapSystemStartDiffOutput struct {
	Started bool `json:"started"`
}

type sapSystemStateChangeArguments struct {
	instNumber   string
	timeout      time.Duration
	instanceType sapcontrolapi.StartStopOption
}

type SAPSystemStartOption Option[SAPSystemStart]

// SAPSystemStart operator starts a SAP system.
//
// Arguments:
//  instance_number (required): String with the instance number of local instance to start the whole system
//  timeout: Timeout in seconds to wait until the system is started
//  instance_type: Instance type to filter in the StartSystem process. Values: all|abap|j2ee|scs|enqrep.
//  Default value: all
//
// # Execution Phases
//
// - PLAN:
//   The operator gets the system current instances and stores the state.
//   The operation is skipped if the SAP system is already started.
//
// - COMMIT:
//   It starts the SAP system using the sapcontrol StartSystem command.
//
// - VERIFY:
//   Verify if the SAP system is started.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the system is stopped back again.

type SAPSystemStart struct {
	baseOperator
	parsedArguments     *sapSystemStateChangeArguments
	sapControlConnector sapcontrolapi.WebService
	interval            time.Duration
}

func WithCustomStartSystemSapcontrol(sapControlConnector sapcontrolapi.WebService) SAPSystemStartOption {
	return func(o *SAPSystemStart) {
		o.sapControlConnector = sapControlConnector
	}
}

func WithCustomStartSystemInterval(interval time.Duration) SAPSystemStartOption {
	return func(o *SAPSystemStart) {
		o.interval = interval
	}
}

func NewSAPSystemStart(
	arguments Arguments,
	operationID string,
	options Options[SAPSystemStart],
) *Executor {
	sapSystemStart := &SAPSystemStart{
		baseOperator: newBaseOperator(
			SapSystemStartOperatorName, operationID, arguments, options.BaseOperatorOptions...,
		),
		interval: defaultSapSystemStateInterval,
	}

	for _, opt := range options.OperatorOptions {
		opt(sapSystemStart)
	}

	return &Executor{
		phaser:      sapSystemStart,
		operationID: operationID,
		logger:      sapSystemStart.logger,
	}
}

func (s *SAPSystemStart) plan(ctx context.Context) (bool, error) {
	opArguments, err := parseSAPSystemStateChangeArguments(s.arguments)
	if err != nil {
		return false, err
	}
	s.parsedArguments = opArguments

	// Use custom sapControlConnector or create a new one based on the instance_number argument
	if s.sapControlConnector == nil {
		s.sapControlConnector = sapcontrolapi.NewWebServiceUnix(s.parsedArguments.instNumber)
	}

	started, err := allInstancesInState(
		ctx,
		s.sapControlConnector,
		s.parsedArguments.instanceType,
		sapcontrolapi.STATECOLOR_GREEN,
	)
	if err != nil {
		return false, fmt.Errorf("error checking system state: %w", err)
	}

	s.resources[beforeDiffField] = started

	if started {
		s.logger.Info("system already started, skipping operation")
		s.resources[afterDiffField] = started
		return true, nil
	}

	return false, nil
}

func (s *SAPSystemStart) commit(ctx context.Context) error {
	request := new(sapcontrolapi.StartSystem)
	request.Options = &s.parsedArguments.instanceType
	// Even though the timeout is optional, the action doesn't work if a value is not set
	// and `all` instance_type is used. It is sent as 0, which means that it
	// would hit the timeout and stop the operation
	request.Waittimeout = int32(s.parsedArguments.timeout.Seconds())
	_, err := s.sapControlConnector.StartSystemContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting system: %w", err)
	}

	return nil
}

func (s *SAPSystemStart) verify(ctx context.Context) error {
	err := waitUntilSapSystemState(
		ctx,
		s.sapControlConnector,
		s.parsedArguments.instanceType,
		sapcontrolapi.STATECOLOR_GREEN,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("verify system started failed: %w", err)
	}

	s.resources[afterDiffField] = true
	return nil
}

func (s *SAPSystemStart) rollback(ctx context.Context) error {
	request := new(sapcontrolapi.StopSystem)
	request.Options = &s.parsedArguments.instanceType
	_, err := s.sapControlConnector.StopSystemContext(ctx, request)
	if err != nil {
		return fmt.Errorf("error stopping system: %w", err)
	}

	err = waitUntilSapSystemState(
		ctx,
		s.sapControlConnector,
		s.parsedArguments.instanceType,
		sapcontrolapi.STATECOLOR_GRAY,
		s.parsedArguments.timeout,
		s.interval,
	)

	if err != nil {
		return fmt.Errorf("rollback to stopped failed: %w", err)
	}

	return nil
}

//	operationDiff needs to be refactored, ignoring duplication issues for now
//
// nolint: dupl
func (s *SAPSystemStart) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeStarted, ok := s.resources[beforeDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid beforeStarted value: cannot parse '%s' to bool",
			s.resources[beforeDiffField]))
	}

	afterStarted, ok := s.resources[afterDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid afterStarted value: cannot parse '%s' to bool",
			s.resources[afterDiffField]))
	}

	beforeDiffOutput := sapSystemStartDiffOutput{
		Started: beforeStarted,
	}
	before, err := json.Marshal(beforeDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling before diff output: %v", err))
	}
	diff["before"] = string(before)

	afterDiffOutput := sapSystemStartDiffOutput{
		Started: afterStarted,
	}
	after, err := json.Marshal(afterDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling after diff output: %v", err))
	}
	diff["after"] = string(after)

	return diff
}

func allInstancesInState(
	ctx context.Context,
	connector sapcontrolapi.WebService,
	instanceType sapcontrolapi.StartStopOption,
	expectedState sapcontrolapi.STATECOLOR,
) (bool, error) {
	request := new(sapcontrolapi.GetSystemInstanceList)
	response, err := connector.GetSystemInstanceListContext(ctx, request)
	if err != nil {
		return false, fmt.Errorf("error getting instance list: %w", err)
	}

	filteringMap := map[sapcontrolapi.StartStopOption]string{
		sapcontrolapi.StartStopOptionSAPControlALLINSTANCES:    "",
		sapcontrolapi.StartStopOptionSAPControlABAPINSTANCES:   "ABAP",
		sapcontrolapi.StartStopOptionSAPControlJ2EEINSTANCES:   "J2EE",
		sapcontrolapi.StartStopOptionSAPControlSCSINSTANCES:    "MESSAGESERVER",
		sapcontrolapi.StartStopOptionSAPControlENQREPINSTANCES: "ENQREP",
	}
	filteringValue := filteringMap[instanceType]

	for _, instance := range response.Instances {
		// filter out instances that are not part of the current instance type value
		if !strings.Contains(instance.Features, filteringValue) {
			continue
		}

		if instance.Dispstatus != expectedState {
			return false, nil
		}
	}

	return true, nil
}

func waitUntilSapSystemState(
	ctx context.Context,
	connector sapcontrolapi.WebService,
	instanceType sapcontrolapi.StartStopOption,
	expectedState sapcontrolapi.STATECOLOR,
	timeout time.Duration,
	interval time.Duration,
) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		inState, err := allInstancesInState(timeoutCtx, connector, instanceType, expectedState)
		if err != nil {
			return err
		}

		if timeoutCtx.Err() != nil {
			return fmt.Errorf("error waiting until system is in desired state")
		}

		if inState {
			return nil
		}

		err = sleepContext(timeoutCtx, interval)
		if err != nil {
			return err
		}
	}

}

func parseSAPSystemStateChangeArguments(rawArguments Arguments) (*sapSystemStateChangeArguments, error) {
	instNumberArgument, found := rawArguments["instance_number"]
	if !found {
		return nil, fmt.Errorf("argument instance_number not provided, could not use the operator")
	}

	instNumber, ok := instNumberArgument.(string)
	if !ok {
		return nil, fmt.Errorf(
			"could not parse instance_number argument as string, argument provided: %v",
			instNumberArgument,
		)
	}

	timeout := defaultSapSystemStateTimeout
	if timeoutArgument, found := rawArguments["timeout"]; found {
		timeoutFloat, ok := timeoutArgument.(float64)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse timeout argument as a number, argument provided: %v",
				timeoutArgument,
			)
		}

		timeout = time.Duration(timeoutFloat) * time.Second
	}

	instanceType := defaultSapSystemStateInstanceType
	if instanceTypeArgument, found := rawArguments["instance_type"]; found {
		instanceTypeStr, ok := instanceTypeArgument.(string)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse instance_type argument as a string, argument provided: %v",
				instanceTypeArgument,
			)
		}

		instancesMap := map[string]sapcontrolapi.StartStopOption{
			instanceTypeALL:    sapcontrolapi.StartStopOptionSAPControlALLINSTANCES,
			instanceTypeABAP:   sapcontrolapi.StartStopOptionSAPControlABAPINSTANCES,
			instanceTypeJ2EE:   sapcontrolapi.StartStopOptionSAPControlJ2EEINSTANCES,
			instanceTypeSCS:    sapcontrolapi.StartStopOptionSAPControlSCSINSTANCES,
			instanceTypeENQREP: sapcontrolapi.StartStopOptionSAPControlENQREPINSTANCES,
		}
		instanceType, ok = instancesMap[instanceTypeStr]
		if !ok {
			return nil, fmt.Errorf("invalid instance_type value: %s", instanceTypeStr)
		}
	}

	return &sapSystemStateChangeArguments{
		instNumber:   instNumber,
		timeout:      timeout,
		instanceType: instanceType,
	}, nil
}
