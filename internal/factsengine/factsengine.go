package factsengine

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/adapters"
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
	gathererManager     gatherers.Manager
	factsServiceAdapter adapters.Adapter
	pluginLoaders       PluginLoaders
}

func NewFactsEngine(agentID, factsEngineService string, manager gatherers.Manager) *FactsEngine {
	return &FactsEngine{
		agentID:             agentID,
		factsEngineService:  factsEngineService,
		factsServiceAdapter: nil,
		gathererManager:     manager,
		pluginLoaders:       NewPluginLoaders(),
	}
}

func mergeGatherers(maps ...map[string]gatherers.FactGatherer) map[string]gatherers.FactGatherer {
	result := make(map[string]gatherers.FactGatherer)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// Already in the manager, populate in the main

// func (c *FactsEngine) LoadPlugins(pluginsFolder string) error {
// 	loadedPlugins, err := loadPlugins(c.pluginLoaders, pluginsFolder)
// 	if err != nil {
// 		return errors.Wrap(err, "Error loading plugins")
// 	}

// 	allGatherers := mergeGatherers(c.factGatherers, loadedPlugins)
// 	c.factGatherers = allGatherers

// 	return nil
// }

func (c *FactsEngine) CleanupPlugins() {
	cleanupPlugins()
}

func (c *FactsEngine) Subscribe() error {
	log.Infof("Subscribing agent %s to the facts gathering reception service on %s", c.agentID, c.factsEngineService)
	// RabbitMQ adapter exists only by now
	factsServiceAdapter, err := adapters.NewRabbitMQAdapter(c.factsEngineService)
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
		c.CleanupPlugins()
		err = c.Unsubscribe()
		if err != nil {
			log.Errorf("Error during unsubscription: %s", err)
		}
	}()
	queue := fmt.Sprintf(agentsQueue, c.agentID)
	if err := c.factsServiceAdapter.Listen(queue, exchange, agentsEventsRoutingKey, c.handleEvent); err != nil {
		return err
	}

	<-ctx.Done()

	return err
}
