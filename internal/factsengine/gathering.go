package factsengine

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"golang.org/x/sync/errgroup"
)

func gatherFacts(
	agentID string,
	factsRequest gatherers.FactsRequest,
	factGatherers map[string]gatherers.FactGatherer,
) (gatherers.FactsResult, error) {
	factsResults := gatherers.FactsResult{
		ExecutionID: factsRequest.ExecutionID,
		AgentID:     agentID,
		Facts:       nil,
	}
	groupedFactsRequest := groupFactsRequestByGatherer(factsRequest)
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

// Group the received facts by gatherer type, so they are executed in the same moment with the same source of truth
func groupFactsRequestByGatherer(factsRequest gatherers.FactsRequest) gatherers.GroupedFactsRequest {
	groupedFactsRequest := gatherers.GroupedFactsRequest{
		ExecutionID: factsRequest.ExecutionID,
		Facts:       make(map[string][]gatherers.FactRequest),
	}

	for _, factRequest := range factsRequest.Facts {
		groupedFactsRequest.Facts[factRequest.Gatherer] = append(groupedFactsRequest.Facts[factRequest.Gatherer], factRequest)
	}

	return groupedFactsRequest
}
