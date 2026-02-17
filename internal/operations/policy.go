package operations

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/trento-project/agent/internal/messaging"

	"github.com/trento-project/agent/internal/operations/operator"
	"github.com/trento-project/contracts/go/pkg/events"
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
		return fmt.Errorf("error getting event type: %w", err)
	}
	switch eventType {
	case OperatorExecutionRequestedV1:
		operatorExecutionRequested, err := OperatorExecutionRequestedFromEvent(event)
		if err != nil {
			return fmt.Errorf("error decoding OperatorExecutionRequested event: %w", err)
		}
		slog.Info("Operator execution request received", "operator", operatorExecutionRequested.Operator)

		target := operatorExecutionRequested.GetTargetAgent(agentID)
		if target == nil {
			slog.Info("OperatorExecutionRequested is not for this agent. Discarding operator execution")
			return nil
		}

		operatorBuilder, err := registry.GetOperatorBuilder(operatorExecutionRequested.Operator)
		if err != nil {
			return fmt.Errorf("error building operator from operators registry: %w", err)
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
			return fmt.Errorf("error encoding OperatorExecutionCompleted event: %w", err)
		}

		slog.Info("Operator execution request completed", "operator", operatorExecutionRequested.Operator)

		if err := adapter.Publish(
			operationsRoutingKey, events.ContentType(), completedEvent); err != nil {
			return fmt.Errorf("error publishing operator execution report: %w", err)
		}

		slog.Info("Operation report published properly")

		return nil
	default:
		return fmt.Errorf("invalid event type: %s", eventType)
	}
}
