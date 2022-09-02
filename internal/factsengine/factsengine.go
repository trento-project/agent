package factsengine

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/adapters"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"golang.org/x/sync/errgroup"
)

const (
	gatherFactsExchange string = "gather_facts"
	factsExchange       string = "facts"
)

type FactsEngine struct {
	agentID             string
	factsEngineService  string
	factGatherers       map[string]gatherers.FactGatherer
	factsServiceAdapter adapters.Adapter
	pluginLoaders       PluginLoaders
}

func NewFactsEngine(agentID, factsEngineService string) *FactsEngine {
	return &FactsEngine{
		agentID:             agentID,
		factsEngineService:  factsEngineService,
		factsServiceAdapter: nil,
		factGatherers: map[string]gatherers.FactGatherer{
			gatherers.CorosyncFactKey:            gatherers.NewDefaultCorosyncConfGatherer(),
			gatherers.CorosyncCmapCtlFactKey:     gatherers.NewDefaultCorosyncCmapctlGatherer(),
			gatherers.PackageVersionGathererName: gatherers.NewDefaultPackageVersionGatherer(),
			gatherers.CrmMonGathererName:         gatherers.NewDefaultCrmMonGatherer(),
			gatherers.CibAdminGathererName:       gatherers.NewDefaultCibAdminGatherer(),
			gatherers.SystemDGathererName:        gatherers.NewSystemDGatherer(),
			gatherers.SBDConfigGathererName:      gatherers.NewSBDGathererWithDefaultConfig(),
			gatherers.VerifyPasswordGathererName: gatherers.NewDefaultPasswordGatherer(),
		},
		pluginLoaders: NewPluginLoaders(),
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

func (c *FactsEngine) LoadPlugins(pluginsFolder string) error {
	loadedPlugins, err := loadPlugins(c.pluginLoaders, pluginsFolder)
	if err != nil {
		return errors.Wrap(err, "Error loading plugins")
	}

	allGatherers := mergeGatherers(c.factGatherers, loadedPlugins)
	c.factGatherers = allGatherers

	return nil
}

func (c *FactsEngine) CleanupPlugins() {
	cleanupPlugins()
}

func (c *FactsEngine) GetGatherer(gatherer string) (gatherers.FactGatherer, error) {
	if g, found := c.factGatherers[gatherer]; found {
		return g, nil
	}
	return nil, errors.Errorf("gatherer %s not found", gatherer)
}

func (c *FactsEngine) GetGatherersList() []string {
	gatherersList := []string{}

	for gatherer := range c.factGatherers {
		gatherersList = append(gatherersList, gatherer)
	}

	return gatherersList
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

	if err := c.factsServiceAdapter.Listen(c.agentID, gatherFactsExchange, c.handleRequest); err != nil {
		return err
	}

	<-ctx.Done()

	return err
}

func (c *FactsEngine) handleRequest(contentType string, request []byte) error {
	factsRequests, err := parseFactsRequest(request)
	if err != nil {
		log.Errorf("Invalid facts request: %s", err)
		return err
	}

	gatheredFacts, err := gatherFacts(c.agentID, factsRequests, c.factGatherers)
	if err != nil {
		log.Errorf("Error gathering facts: %s", err)
		return err
	}

	if err := c.publishFacts(gatheredFacts); err != nil {
		log.Errorf("Error publishing facts: %s", err)
		return err
	}

	return nil
}

func gatherFacts(
	agentID string,
	groupedFactsRequest *gatherers.GroupedFactsRequest,
	factGatherers map[string]gatherers.FactGatherer,
) (gatherers.FactsResult, error) {
	factsResults := gatherers.FactsResult{
		ExecutionID: groupedFactsRequest.ExecutionID,
		AgentID:     agentID,
		Facts:       nil,
	}
	factsCh := make(chan []gatherers.Fact, len(groupedFactsRequest.Facts))
	g := new(errgroup.Group)

	log.Infof("Starting facts gathering process")

	// Gather facts asynchronously
	for gathererType, f := range groupedFactsRequest.Facts {
		factsRequest := f

		gatherer, exists := factGatherers[gathererType]
		if !exists {
			log.Errorf("Fact gatherer %s does not exist", gathererType)
			continue
		}

		// Execute the fact gathering asynchronously and in parallel
		g.Go(func() error {
			if newFacts, err := gatherer.Gather(factsRequest); err == nil {
				factsCh <- newFacts
			} else {
				log.Error(err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return factsResults, err
	}

	close(factsCh)

	for newFacts := range factsCh {
		factsResults.Facts = append(factsResults.Facts, newFacts...)
	}

	log.Infof("Requested facts gathered")
	return factsResults, nil
}

func parseFactsRequest(request []byte) (*gatherers.GroupedFactsRequest, error) {
	var factsRequest gatherers.FactsRequest
	var groupedFactsRequest *gatherers.GroupedFactsRequest

	err := json.Unmarshal(request, &factsRequest)
	if err != nil {
		return nil, err
	}

	groupedFactsRequest = &gatherers.GroupedFactsRequest{
		ExecutionID: factsRequest.ExecutionID,
		Facts:       make(map[string][]gatherers.FactRequest),
	}

	// Group the received facts by gatherer type, so they are executed in the same moment with the same source of truth
	for _, factRequest := range factsRequest.Facts {
		groupedFactsRequest.Facts[factRequest.Gatherer] = append(groupedFactsRequest.Facts[factRequest.Gatherer], factRequest)
	}

	return groupedFactsRequest, nil
}
