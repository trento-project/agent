package gatherers

import "github.com/trento-project/agent/internal/factsengine/entities"

type FactGatherer interface {
	Gather(factsRequests []entities.FactRequest) ([]entities.FactsGatheredItem, error)
}
