package operations

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/trento-project/contracts/go/pkg/events"
	"github.com/trento-project/workbench/pkg/operator"
)

const (
	EventSource = "https://github.com/trento-project/agent"
)

type OperatorExecutionRequestedTarget struct {
	AgentID   string
	Arguments map[string]interface{}
}

type OperatorExecutionRequested struct {
	OperationID string
	GroupID     string
	StepNumber  int32
	Operator    string
	Targets     []OperatorExecutionRequestedTarget
}

func (r *OperatorExecutionRequested) GetTargetAgent(agentID string) *OperatorExecutionRequestedTarget {
	var currentAgent *OperatorExecutionRequestedTarget

	for _, target := range r.Targets {
		if target.AgentID == agentID {
			currentAgent = &target
		}
	}

	return currentAgent
}

func OperatorExecutionRequestedFromEvent(event []byte) (*OperatorExecutionRequested, error) {
	var operatorExecutionRequested events.OperatorExecutionRequested

	err := events.FromEvent(event, &operatorExecutionRequested, events.WithExpirationCheck())
	if err != nil {
		return nil, err
	}

	targets := []OperatorExecutionRequestedTarget{}

	for _, target := range operatorExecutionRequested.Targets {
		arguments := make(map[string]interface{})
		for key, value := range target.GetArguments() {
			arguments[key] = value.AsInterface()
		}

		newTarget := OperatorExecutionRequestedTarget{
			AgentID:   target.GetAgentId(),
			Arguments: arguments,
		}

		targets = append(targets, newTarget)
	}

	return &OperatorExecutionRequested{
		OperationID: operatorExecutionRequested.GetOperationId(),
		GroupID:     operatorExecutionRequested.GetGroupId(),
		StepNumber:  operatorExecutionRequested.GetStepNumber(),
		Operator:    operatorExecutionRequested.GetOperator(),
		Targets:     targets,
	}, nil
}

func OperatorExecutionCompletedToEvent(
	operationID,
	groupID,
	agentID string,
	stepNumber int32,
	report *operator.ExecutionReport,
) ([]byte, error) {
	event := events.OperatorExecutionCompleted{
		OperationId: operationID,
		GroupId:     groupID,
		StepNumber:  stepNumber,
		AgentId:     agentID,
	}

	if report.Success != nil {
		before, beforeFound := report.Success.Diff["before"]
		if !beforeFound {
			return nil, fmt.Errorf("before not found in report")
		}

		beforeValue, err := structpb.NewValue(before)
		if err != nil {
			return nil, err
		}

		after, afterFound := report.Success.Diff["after"]
		if !afterFound {
			return nil, fmt.Errorf("after not found in report")
		}

		afterValue, err := structpb.NewValue(after)
		if err != nil {
			return nil, err
		}

		result := &events.OperatorExecutionCompleted_Value{
			Value: &events.OperatorResponse{
				Phase: events.OperatorPhase(events.OperatorPhase_value[string(report.Success.LastPhase)]),
				Diff: &events.OperatorDiff{
					Before: beforeValue,
					After:  afterValue,
				},
			},
		}
		event.Result = result
	} else {
		result := &events.OperatorExecutionCompleted_Error{
			Error: &events.OperatorError{
				Phase:   events.OperatorPhase(events.OperatorPhase_value[string(report.Error.ErrorPhase)]),
				Message: report.Error.Message,
			},
		}
		event.Result = result
	}

	eventBytes, err := events.ToEvent(
		&event,
		events.WithSource(EventSource),
		events.WithID(uuid.New().String()),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating event: %w", err)
	}

	return eventBytes, nil
}
