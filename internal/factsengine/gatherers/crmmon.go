package gatherers

import (
	log "github.com/sirupsen/logrus"
)

const (
	CrmMonGathererName = "crm_mon"
)

type CrmMonGatherer struct {
	executor CommandExecutor
}

func NewCrmMonGatherer() *CrmMonGatherer {
	return &CrmMonGatherer{
		executor: Executor{},
	}
}

func (g *CrmMonGatherer) Gather(factsRequests []FactRequest) ([]Fact, error) {
	log.Infof("Starting crmmon facts gathering process")

	crmmon, err := g.executor.Exec("crm_mon", "--output-as", "xml")
	if err != nil {
		return nil, err
	}

	facts, err := GatherFromXML(string(crmmon), factsRequests)

	log.Infof("Requested crmmon facts gathered")
	return facts, err
}
