package factsengine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/messaging"
)

const (
	exchange               string = "trento.checks"
	agentsQueue            string = "trento.checks.agents.%s"
	agentsEventsRoutingKey string = "agents"
	executionsRoutingKey   string = "executions"
)

type FactsEngine struct {
	agentID             string
	factsEngineService  string
	gathererRegistry    gatherers.Registry
	factsServiceAdapter messaging.Adapter
}

func NewFactsEngine(agentID, factsEngineService string, registry gatherers.Registry) *FactsEngine {
	return &FactsEngine{
		agentID:             agentID,
		factsEngineService:  factsEngineService,
		factsServiceAdapter: nil,
		gathererRegistry:    registry,
	}
}

func (c *FactsEngine) Subscribe() error {
	slog.Info("Subscribing agent to the facts gathering reception service",
		"agentID", c.agentID, "factsEngineService", c.factsEngineService)
	// RabbitMQ adapter exists only by now
	queue := fmt.Sprintf(agentsQueue, c.agentID)
	factsServiceAdapter, err := messaging.NewRabbitMQAdapter(
		c.factsEngineService,
		queue,
		exchange,
		agentsEventsRoutingKey,
	)
	if err != nil {
		return err
	}

	c.factsServiceAdapter = factsServiceAdapter
	slog.Info("Subscription to the facts engine by agent done",
		"agentID", c.agentID, "factsEngineService", c.factsEngineService)

	return nil
}

func (c *FactsEngine) Unsubscribe() error {
	slog.Info("Unsubscribing agent from the facts engine service", "agentID", c.agentID)
	if err := c.factsServiceAdapter.Unsubscribe(); err != nil {
		return err
	}

	slog.Info("Unsubscribed properly")

	return nil
}

func (c *FactsEngine) Listen(ctx context.Context) error {
	var err error

	slog.Info("Listening for facts gathering events...")
	defer func() {
		gatherers.CleanupPlugins()
		err = c.Unsubscribe()
		if err != nil {
			slog.Error("Error during unsubscription", "error", err.Error())
		}
	}()
	eventHandler := messaging.MakeEventHandler(
		ctx,
		c.agentID,
		c.factsServiceAdapter,
		c.gathererRegistry,
		HandleEvent,
	)

	if err := c.factsServiceAdapter.Listen(eventHandler); err != nil {
		return err
	}

	<-ctx.Done()

	return err
}
