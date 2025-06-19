package operations

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/trento-project/agent/internal/messaging"
	"github.com/trento-project/workbench/pkg/operator"
)

const (
	exchange               string = "trento.operations"
	agentsQueue            string = "trento.operations.agents.%s"
	agentsEventsRoutingKey string = "agents"
	operationsRoutingKey   string = "requests"
)

type Engine struct {
	agentID          string
	amqpServiceURL   string
	amqpAdapter      messaging.Adapter
	operatorRegistry operator.Registry
}

func NewOperationsEngine(agentID, amqpServiceURL string, registry operator.Registry) *Engine {
	return &Engine{
		agentID:          agentID,
		amqpServiceURL:   amqpServiceURL,
		amqpAdapter:      nil,
		operatorRegistry: registry,
	}
}

func (e *Engine) Subscribe() error {
	slog.Info("Subscribing agent to the operations reception service",
		"agent_id", e.agentID,
		"amqp_service_url", e.amqpServiceURL)
	queue := fmt.Sprintf(agentsQueue, e.agentID)
	amqpAdapter, err := messaging.NewRabbitMQAdapter(
		e.amqpServiceURL,
		queue,
		exchange,
		agentsEventsRoutingKey,
	)
	if err != nil {
		return err
	}

	e.amqpAdapter = amqpAdapter
	slog.Info("Subscription to the operations engine by agent done",
		"agent_id", e.agentID,
		"amqp_service_url", e.amqpServiceURL)

	return nil
}

func (e *Engine) Unsubscribe() error {
	slog.Info("Unsubscribing agent from the operations engine service", "agent_id", e.agentID)
	if err := e.amqpAdapter.Unsubscribe(); err != nil {
		return err
	}

	slog.Info("Unsubscribed properly")

	return nil
}

func (e *Engine) Listen(ctx context.Context) error {
	var err error

	slog.Info("Listening for operation events...")
	defer func() {
		err = e.Unsubscribe()
		if err != nil {
			slog.Error("Error during unsubscription", "error", err)
		}
	}()
	eventHandler := messaging.MakeEventHandler(
		ctx,
		e.agentID,
		e.amqpAdapter,
		e.operatorRegistry,
		HandleEvent,
	)
	if err := e.amqpAdapter.Listen(eventHandler); err != nil {
		return err
	}

	<-ctx.Done()

	return err
}
