package factsengine

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/messaging"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
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
	log.Infof("Subscribing agent %s to the facts gathering reception service on %s", c.agentID, c.factsEngineService)
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
	log.Infof("Subscription to the facts engine by agent %s in %s done", c.agentID, c.factsEngineService)

	return nil
}

func (c *FactsEngine) Unsubscribe() error {
	log.Infof("Unsubscribing agent %s from the facts engine service", c.agentID)
	if err := c.factsServiceAdapter.Unsubscribe(); err != nil {
		return err
	}

	log.Infof("Unsubscribed properly")

	return nil
}

func (c *FactsEngine) Listen(ctx context.Context) error {
	var err error

	log.Infof("Listening for facts gathering events...")
	defer func() {
		gatherers.CleanupPlugins()
		err = c.Unsubscribe()
		if err != nil {
			log.Errorf("Error during unsubscription: %s", err)
		}
	}()
	if err := c.factsServiceAdapter.Listen(c.makeEventHandler(ctx)); err != nil {
		return err
	}

	<-ctx.Done()

	return err
}
