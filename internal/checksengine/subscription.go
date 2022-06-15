package checksengine

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/trento-project/agent/internal/checksengine/facts"
)

type checksEngine struct {
	agentID             string
	checksEngineService string
	gatherers           map[string]facts.FactGatherer
}

func NewChecksEngine(agentID, checksEngineService string) *checksEngine {
	return &checksEngine{
		agentID:             agentID,
		checksEngineService: checksEngineService,
		gatherers: map[string]facts.FactGatherer{
			facts.SBDFactKey: facts.NewSbdConfigGatherer(),
		},
	}
}

func (c *checksEngine) Subscribe() error {
	log.Infof("Subscribing agent %s to the checks engine runner on %s", c.agentID, c.checksEngineService)
	// Subscribe somehow to the checks engine runner
	log.Infof("Subscription to the checks engine by agent %s in %s done", c.agentID, c.checksEngineService)

	return nil
}

func (c *checksEngine) Unsubscribe() error {
	log.Infof("Unsubscribing agent %s to the checks engine runner", c.agentID)
	// Unsubscribe somehow from the checks engine runner
	log.Infof("Unsubscribed properly")

	return nil

}

func (c *checksEngine) Listen(ctx context.Context) {
	log.Infof("Listening for checks execution events...")
	defer c.Unsubscribe()

	// Dummy code to gather SBD configuration files every some seconds
	c.dummyGatherer(ctx)
}

func (c *checksEngine) dummyGatherer(ctx context.Context) {
	var factsRequests = []*facts.FactsRequest{
		{
			Name: facts.SBDFactKey,
			Keys: []string{"SDB_DEVICE", "SBD_TIMEOUT_ACTION"},
		},
	}
	ticker := time.NewTicker(10 * time.Second)

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			gatheredFacts, _ := gatherFacts(factsRequests, c.gatherers)
			publishFacts(gatheredFacts)
		case <-ctx.Done():
			return
		}
	}
}

func gatherFacts(factsRequests []*facts.FactsRequest, gatherers map[string]facts.FactGatherer) ([]*facts.Fact, error) {
	var gatheredFacts []*facts.Fact
	log.Infof("Starting facts gathering process")

	// Gather facts asynchronously
	var wg sync.WaitGroup

	for _, factRequest := range factsRequests {

		g, exists := gatherers[factRequest.Name]
		if !exists {
			log.Errorf("Fact gatherer %s does not exist", factRequest.Name)
			continue
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, factRequestKeys []string) {
			defer wg.Done()
			newFacts, err := g.Gather(factRequestKeys)
			if err == nil {
				gatheredFacts = append(gatheredFacts, newFacts[:]...)
			}
		}(&wg, factRequest.Keys)
	}

	wg.Wait()

	log.Infof("Requested facts gathered")
	return gatheredFacts, nil
}

func buildResponse(facts []*facts.Fact) ([]byte, error) {
	log.Infof("Building gathered facts response...")

	jsonFacts, err := json.Marshal(facts)
	if err != nil {
		return nil, err
	}

	log.Infof("Gathered facts response built properly")

	return jsonFacts, nil
}

func publishFacts(facts []*facts.Fact) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	response, err := buildResponse(facts)
	if err != nil {
		log.Errorf("Error building response: %v", err)
		return err
	}

	// By now, simply print the gathered facts
	log.Infof("Gathered facts response: %s", response)

	// Publish somehow the gathered facts
	log.Infof("Gathered facts published properly")
	return nil
}
