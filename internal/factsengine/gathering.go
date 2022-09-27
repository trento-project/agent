package factsengine

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"golang.org/x/sync/errgroup"
)

func gatherFacts(
	executionID,
	agentID string,
	agentFacts *entities.FactsGatheringRequestedTarget,
	registry gatherers.Registry,
) (entities.FactsGathered, error) {
	factsResults := entities.FactsGathered{
		ExecutionID:   executionID,
		AgentID:       agentID,
		FactsGathered: nil,
	}
	groupedFactsRequest := groupFactsRequestByGatherer(agentFacts)
	factsCh := make(chan []entities.Fact, len(groupedFactsRequest.FactRequests))
	g := new(errgroup.Group)

	log.Infof("Starting facts gathering process")

	// Gather facts asynchronously
	for gathererType, f := range groupedFactsRequest.FactRequests {
		factsRequest := f

		gatherer, err := registry.GetGatherer(gathererType)
		if err != nil {
			log.Errorf("Fact gatherer %s does not exist", gathererType)
			continue
		}

		// Execute the fact gathering asynchronously and in parallel
		g.Go(func() error {
			var gatheringError *entities.FactGatheringError

			newFacts, err := gatherer.Gather(factsRequest)
			switch {
			case err == nil:
				factsCh <- newFacts
			case errors.As(err, &gatheringError):
				log.Error(gatheringError)
				factsCh <- entities.NewFactsGatheredListWithError(factsRequest, gatheringError)
			default:
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
		factsResults.FactsGathered = append(factsResults.FactsGathered, newFacts...)
	}

	log.Infof("Requested facts gathered")
	return factsResults, nil
}

// Group the received facts by gatherer type, so they are executed in the same moment with the same source of truth
func groupFactsRequestByGatherer(
	factsRequest *entities.FactsGatheringRequestedTarget) entities.GroupedByGathererRequestedTarget {

	groupedFactsRequest := entities.GroupedByGathererRequestedTarget{
		FactRequests: make(map[string][]entities.FactRequest),
	}

	for _, factRequest := range factsRequest.FactRequests {
		groupedFactsRequest.FactRequests[factRequest.Gatherer] = append(
			groupedFactsRequest.FactRequests[factRequest.Gatherer], factRequest)
	}

	return groupedFactsRequest
}
