package operations

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/trento-project/agent/internal/messaging"

	"github.com/trento-project/contracts/go/pkg/events"
	"github.com/trento-project/workbench/pkg/operator"
)

const (
	OperatorExecutionRequestedV1 = "Trento.Operations.V1.OperatorExecutionRequested"
)

func HandleEvent(
	ctx context.Context,
	event []byte,
	agentID string,
	adapter messaging.Adapter,
	registry operator.Registry,
) error {
	eventType, err := events.EventType(event)
	if err != nil {
		return errors.Wrap(err, "error getting event type")
	}
	switch eventType {
	case OperatorExecutionRequestedV1:
		operatorExecutionRequested, err := OperatorExecutionRequestedFromEvent(event)
		if err != nil {
			return errors.Wrap(err, "error decoding OperatorExecutionRequested event")
		}
		slog.Info("Operator execution request received", "operator", operatorExecutionRequested.Operator)

		target := operatorExecutionRequested.GetTargetAgent(agentID)
		if target == nil {
			slog.Info("OperatorExecutionRequested is not for this agent. Discarding operator execution")
			return nil
		}

		operatorBuilder, err := registry.GetOperatorBuilder(operatorExecutionRequested.Operator)
		if err != nil {
			return errors.Wrap(err, "error building operator from operators registry")
		}
		op := operatorBuilder(operatorExecutionRequested.OperationID, target.Arguments)
		report := op.Run(ctx)

		completedEvent, err := OperatorExecutionCompletedToEvent(
			operatorExecutionRequested.OperationID,
			operatorExecutionRequested.GroupID,
			target.AgentID,
			operatorExecutionRequested.StepNumber,
			report,
		)
		if err != nil {
			return errors.Wrap(err, "error encoding OperatorExecutionCompleted event")
		}

		slog.Info("Operator execution request completed", "operator", operatorExecutionRequested.Operator)

		if err := adapter.Publish(
			operationsRoutingKey, events.ContentType(), completedEvent); err != nil {
			return errors.Wrap(err, "error publishing operator execution report")
		}

		slog.Info("Operation report published properly")

		return nil
	default:
		return fmt.Errorf("invalid event type: %s", eventType)
	}
}
