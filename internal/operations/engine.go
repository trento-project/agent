package operations

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
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
	log.Infof("Subscribing agent %s to the operations reception service on %s", e.agentID, e.amqpServiceURL)
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
	log.Infof("Subscription to the operations engine by agent %s in %s done", e.agentID, e.amqpServiceURL)

	return nil
}

func (e *Engine) Unsubscribe() error {
	log.Infof("Unsubscribing agent %s from the operations engine service", e.agentID)
	if err := e.amqpAdapter.Unsubscribe(); err != nil {
		return err
	}

	log.Infof("Unsubscribed properly")

	return nil
}

func (e *Engine) Listen(ctx context.Context) error {
	var err error

	log.Infof("Listening for operation events...")
	defer func() {
		err = e.Unsubscribe()
		if err != nil {
			log.Errorf("Error during unsubscription: %s", err)
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
