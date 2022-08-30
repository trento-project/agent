package gatherers

import (
	log "github.com/sirupsen/logrus"

	"github.com/trento-project/agent/internal/utils"
)

const (
	CibAdminGathererName = "cibadmin"
)

type CibAdminGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultCibAdminGatherer() *CibAdminGatherer {
	return NewCibAdminGatherer(utils.Executor{})
}

func NewCibAdminGatherer(executor utils.CommandExecutor) *CibAdminGatherer {
	return &CibAdminGatherer{
		executor: executor,
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
