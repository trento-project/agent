package gatherers

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SapHostCtrlGathererName = "saphostctrl"
)

// nolint:gochecknoglobals
var (
	SapHostCtrlCommandError = entities.FactGatheringError{
		Type:    "saphostctrl-cmd-error",
		Message: "error executing saphostctrl command",
	}
)

type SapHostCtrlGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultSapHostCtrlGatherer() *SapHostCtrlGatherer {
	return NewSapHostCtrlGatherer(utils.Executor{})
}

func NewSapHostCtrlGatherer(executor utils.CommandExecutor) *SapHostCtrlGatherer {
	return &SapHostCtrlGatherer{
		executor: executor,
	}
}

func (g *SapHostCtrlGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting saphostctrl facts gathering process")

	for _, factReq := range factsRequests {
		var fact entities.Fact
		// TODO: The instance number "00" could be an additional argument in the future
		version, err := g.executor.Exec(
			"/usr/sap/hostctrl/exe/saphostctrl", "-nr", "00", "-function", factReq.Argument)
		if err != nil {
			gatheringError := SapHostCtrlCommandError.Wrap(factReq.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
		} else {
			fact = entities.NewFactGatheredWithRequest(factReq, &entities.FactValueString{Value: (string(version))})
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SapHostCtrlGathererName)
	return facts, nil
}
