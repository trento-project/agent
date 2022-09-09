package gatherers

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

const (
	CrmMonGathererName = "crm_mon"
)

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
		return nil, err
	}

	facts, err := GatherFromXML(string(crmmon), factsRequests)

	log.Infof("Requested crmmon facts gathered")
	return facts, err
}
