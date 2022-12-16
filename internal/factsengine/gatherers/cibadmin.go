package gatherers

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	CibAdminGathererName = "cibadmin"
)

// nolint:gochecknoglobals
var (
	CibAdminCommandError = entities.FactGatheringError{
		Type:    "cibadmin-command-error",
		Message: "error running cibadmin command",
	}

	CibAdminDecodingError = entities.FactGatheringError{
		Type:    "cibadmin-decoding-error",
		Message: "error decoding cibadmin output",
	}
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

func (g *CibAdminGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	log.Infof("Starting %s facts gathering process", CibAdminGathererName)

	cibadmin, err := g.executor.Exec("cibadmin", "--query", "--local")
	if err != nil {
		return nil, CibAdminCommandError.Wrap(err.Error())
	}

	elementsToList := []string{"primitive", "clone", "master", "group",
		"nvpair", "op", "rsc_location", "rsc_order", "rsc_colocation"}

	factValueMap, err := parseXMLToFactValueMap(cibadmin, elementsToList)
	if err != nil {
		return nil, CibAdminDecodingError.Wrap(err.Error())
	}

	facts := []entities.Fact{}

	for _, factReq := range factsRequests {
		var fact entities.Fact

		if value, err := factValueMap.GetValue(factReq.Argument); err == nil {
			fact = entities.NewFactGatheredWithRequest(factReq, value)

		} else {
			log.Error(err)
			fact = entities.NewFactGatheredWithError(factReq, err)
		}
		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", CibAdminGathererName)
	return facts, err
}
