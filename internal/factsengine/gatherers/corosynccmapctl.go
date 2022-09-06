package gatherers

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/internal/utils"
)

const (
	CorosyncCmapCtlFactKey = "corosync-cmapctl"
)

var (
	CmapctlError = entities.FactGatheringError{ // nolint
		Type:    "cmapctl-execution-error",
		Message: "error running cmaptcl command",
	}

	CmapctlValueNotFoundError = entities.FactGatheringError{ // nolint
		Type:    "cmapctl-value-not-found",
		Message: "requested field value not found",
	}
)

type CorosyncCmapctlGatherer struct {
	executor CommandExecutor
}

func NewDefaultCorosyncCmapctlGatherer() *CorosyncCmapctlGatherer {
	return NewCorosyncCmapctlGatherer(Executor{})
}

func NewCorosyncCmapctlGatherer(executor CommandExecutor) *CorosyncCmapctlGatherer {
	return &CorosyncCmapctlGatherer{
		executor: executor,
	}
}

func (s *CorosyncCmapctlGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.FactsGatheredItem, error) {
	facts := []entities.FactsGatheredItem{}
	log.Infof("Starting %s facts gathering process", CorosyncCmapCtlFactKey)

	corosyncCmapctl, err := s.executor.Exec("corosync-cmapctl", "-b")
	if err != nil {
		gatheringError := CmapctlError.Wrap(err.Error())
		log.Errorf(gatheringError.Error())
		return entities.NewFactsGatheredListWithError(factsRequests, &gatheringError), nil
	}

	corosyncCmapctlMap := utils.FindMatches(`(?m)^(\S*)\s\(\S*\)\s=\s(.*)$`, corosyncCmapctl)

	for _, factReq := range factsRequests {
		var fact entities.FactsGatheredItem

		if value, ok := corosyncCmapctlMap[factReq.Argument]; ok {
			fact = entities.NewFactGatheredWithRequest(factReq, fmt.Sprint(value))
		} else {
			gatheringError := CmapctlValueNotFoundError.Wrap(fmt.Sprintf("requested fact %s not found", factReq.Argument))
			log.Errorf(gatheringError.Error())
			fact = entities.NewFactGatheredWithError(factReq, &gatheringError)
		}
		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", CorosyncCmapCtlFactKey)
	return facts, nil
}
