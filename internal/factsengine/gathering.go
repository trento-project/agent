package factsengine

import (
	"context"
	"errors"
	"log/slog"

	"github.com/trento-project/agent/internal/factsengine/factscache"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"golang.org/x/sync/errgroup"
)

func gatherFacts(
	ctx context.Context,
	executionID,
	agentID string,
	groupID string,
	agentFacts *entities.FactsGatheringRequestedTarget,
	registry gatherers.Registry,
) (entities.FactsGathered, error) {
	factsResults := entities.FactsGathered{
		ExecutionID:   executionID,
		AgentID:       agentID,
		FactsGathered: nil,
		GroupID:       groupID,
	}
	groupedFactsRequest := groupFactsRequestByGatherer(agentFacts)
	factsCh := make(chan []entities.Fact, len(groupedFactsRequest.FactRequests))
	cache := factscache.NewFactsCache()

	g := new(errgroup.Group)

	slog.Info("Starting facts gathering process")

	// Gather facts asynchronously
	for gathererType, f := range groupedFactsRequest.FactRequests {
		factsRequest := f

		gatherer, err := registry.GetGatherer(gathererType)
		if err != nil {
			slog.Error("Fact gatherer does not exist", "gathererType", gathererType)
			continue
		}

		// Check if the gatherer implements FactGathererWithCache to set cache
		if gathererWithCache, ok := gatherer.(gatherers.FactGathererWithCache); ok {
			gathererWithCache.SetCache(cache)
		}

		// Execute the fact gathering asynchronously and in parallel
		g.Go(func() error {
			var gatheringError *entities.FactGatheringError

			newFacts, err := gatherer.Gather(ctx, factsRequest)
			ctxErr := ctx.Err()

			switch {
			case ctxErr != nil:
				return ctxErr
			case err == nil:
				factsCh <- newFacts
			case errors.As(err, &gatheringError):
				slog.Error(gatheringError.Error())
				factsCh <- entities.NewFactsGatheredListWithError(factsRequest, gatheringError)
			default:
				slog.Error(err.Error())
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

	slog.Info("Requested facts gathered")
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
