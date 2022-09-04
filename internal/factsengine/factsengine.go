package factsengine

import (
	"context"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/adapters"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	contracts "github.com/trento-project/contracts/go/pkg/gen/entities"
	"golang.org/x/sync/errgroup"
)

const (
	gatherFactsExchange string = "events"
	factsExchange       string = "events"
)

var agentUuid string

type FactsEngine struct {
	agentID             string
	factsEngineService  string
	factGatherers       map[string]gatherers.FactGatherer
	factsServiceAdapter adapters.Adapter
	pluginLoaders       PluginLoaders
}

func NewFactsEngine(agentID, factsEngineService string) *FactsEngine {
	agentUuid = agentID
	return &FactsEngine{
		agentID:             agentID,
		factsEngineService:  factsEngineService,
		factsServiceAdapter: nil,
		factGatherers: map[string]gatherers.FactGatherer{
			gatherers.CorosyncFactKey:            gatherers.NewCorosyncConfGatherer("./test/fixtures/gatherers/corosync.conf.basic"),
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
	event, err := contracts.NewFactsGatheringRequestedV1FromJsonCloudEvent(request)

	if err != nil {
		return err
	}

	agentID := agentUuid

	var interested bool = false
	for _, target := range event.Targets {
		if target == agentID {
			log.Infof("Interested to consumed FactsGatheringRequested even. execution %s", event.ExecutionId)
			interested = true
			break
		}
	}

	if !interested {
		log.Infof("Skipping FactsGatheringRequested for execution %s. Agent %s not among the targets %v", event.ExecutionId, agentID, event.Targets)
		return nil
	}

	factsRequests, err := parseFactsRequest(event)
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

func parseFactsRequest(event *contracts.FactsGatheringRequestedV1) (*gatherers.GroupedFactsRequest, error) {
	var groupedFactsRequest *gatherers.GroupedFactsRequest

	groupedFactsRequest = &gatherers.GroupedFactsRequest{
		ExecutionID: event.ExecutionId,
		Facts:       make(map[string][]gatherers.FactRequest),
	}

	// Group the received facts by gatherer type, so they are executed in the same moment with the same source of truth
	for _, agentFacts := range event.Facts {
		if agentFacts.AgentId == agentUuid {
			for _, agentFact := range agentFacts.Facts {
				groupedFactsRequest.Facts[agentFact.Gatherer] = append(groupedFactsRequest.Facts[agentFact.Gatherer], gatherers.FactRequest{
					Name:     agentFact.Name,
					Gatherer: agentFact.Gatherer,
					Argument: agentFact.Argument,
					CheckID:  agentFact.CheckId,
				})
			}
		}
	}

	return groupedFactsRequest, nil
}
