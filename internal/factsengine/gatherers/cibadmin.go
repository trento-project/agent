package gatherers

import (
	log "github.com/sirupsen/logrus"
)

const (
	CibAdminGathererName = "cibadmin"
)

type CibAdminGatherer struct {
	executor CommandExecutor
}

func NewCibAdminGatherer() *CibAdminGatherer {
	return &CibAdminGatherer{
		executor: Executor{},
	}
}

func (g *CibAdminGatherer) Gather(factsRequests []FactRequest) ([]Fact, error) {
	log.Infof("Starting cibadmin facts gathering process")

	cibadmin, err := g.executor.Exec("cibadmin", "--query", "--local")
	if err != nil {
		return nil, err
	}

	facts, err := GatherFromXML(string(cibadmin), factsRequests)

	log.Infof("Requested cibadmin facts gathered")
	return facts, err
}
