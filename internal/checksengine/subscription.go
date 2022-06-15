package checksengine

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/trento-project/agent/internal/checksengine/facts"
)

func Subscribe(agentID, checksEngineService string) error {
	log.Infof("Subscribing agent %s to the checks engine runner on %s", agentID, checksEngineService)
	// Subscribe somehow to the checks engine runner
	log.Infof("Subscription to the checks engine by agent %s in %s done", agentID, checksEngineService)

	return nil
}

func Unsubscribe(agentID string) error {
	log.Infof("Unsubscribing agent %s to the checks engine runner", agentID)
	// Unsubscribe somehow from the checks engine runner
	log.Infof("Unsubscribed properly")

	return nil

}

func Listen(agentID string, ctx context.Context) {
	log.Infof("Listening for checks execution events...")
	defer Unsubscribe(agentID)

	// Dummy code to gather SBD configuration files every some seconds
	dummyGatherer(ctx)
}

func dummyGatherer(ctx context.Context) {
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
			gatheredFacts, _ := GatherFacts(factsRequests)
			PublishFacts(gatheredFacts)
		case <-ctx.Done():
			return
		}
	}
}

func GatherFacts(factsRequests []*facts.FactsRequest) ([]*facts.Fact, error) {
	var gatheredFacts []*facts.Fact
	log.Infof("Starting facts gathering process")

	// Gather facts asynchronously
	var wg sync.WaitGroup

	for _, factRequest := range factsRequests {
		wg.Add(1)
		switch factRequest.Name {
		case facts.SBDFactKey:
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				sbdConfigFacts, err := facts.GatherSbdConfigFacts(factRequest.Keys)
				if err == nil {
					gatheredFacts = append(gatheredFacts, sbdConfigFacts[:]...)
				}
			}(&wg)
		}
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

func PublishFacts(facts []*facts.Fact) error {
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
