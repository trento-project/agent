package gatherers

import (
	log "github.com/sirupsen/logrus"

	"github.com/trento-project/agent/internal/utils"
)

const (
	CrmMonGathererName = "crm_mon"
)

type CrmMonGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultCrmMonGatherer() *CrmMonGatherer {
	return NewCrmMonGatherer(utils.Executor{})
}

func NewCrmMonGatherer(executor utils.CommandExecutor) *CrmMonGatherer {
	return &CrmMonGatherer{
		executor: executor,
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
