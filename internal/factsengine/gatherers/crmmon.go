package gatherers // nolint

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

const (
	CrmMonGathererName = "crm_mon"
)

var CrmmonError = entities.FactGatheringError{ // nolint
	Type:    "crmmon-execution-error",
	Message: "error running crm_mon command",
}

type CrmMonGatherer struct {
	executor CommandExecutor
}

func NewDefaultCrmMonGatherer() *CrmMonGatherer {
	return NewCrmMonGatherer(Executor{})
}

func NewCrmMonGatherer(executor CommandExecutor) *CrmMonGatherer {
	return &CrmMonGatherer{
		executor: executor,
	}
}

func (g *CrmMonGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.FactsGatheredItem, error) {
	log.Infof("Starting crmmon facts gathering process")

	crmmon, err := g.executor.Exec("crm_mon", "--output-as", "xml")
	if err != nil {
		gatheringError := CrmmonError.Wrap(err.Error())
		log.Errorf(gatheringError.Error())
		return entities.NewFactsGatheredListWithError(factsRequests, &gatheringError), nil
	}

	facts, err := GatherFromXML(string(crmmon), factsRequests)

	log.Infof("Requested crmmon facts gathered")
	return facts, err
}
