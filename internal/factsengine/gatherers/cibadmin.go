package gatherers // nolint

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

const (
	CibAdminGathererName = "cibadmin"
)

var CibadminError = entities.FactGatheringError{ // nolint
	Type:    "cibadmin-execution-error",
	Message: "error running cibadmin command",
}

type CibAdminGatherer struct {
	executor CommandExecutor
}

func NewDefaultCibAdminGatherer() *CibAdminGatherer {
	return NewCibAdminGatherer(Executor{})
}

func NewCibAdminGatherer(executor CommandExecutor) *CibAdminGatherer {
	return &CibAdminGatherer{
		executor: executor,
	}
}

func (g *CibAdminGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.FactsGatheredItem, error) {
	log.Infof("Starting cibadmin facts gathering process")

	cibadmin, err := g.executor.Exec("cibadmin", "--query", "--local")
	if err != nil {
		gatheringError := CibadminError.Wrap(err.Error())
		log.Errorf(gatheringError.Error())
		return entities.NewFactsGatheredListWithError(factsRequests, &gatheringError), nil
	}

	facts, err := GatherFromXML(string(cibadmin), factsRequests)

	log.Infof("Requested cibadmin facts gathered")
	return facts, err
}
