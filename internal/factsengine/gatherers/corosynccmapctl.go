package gatherers

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/utils"
)

const (
	CorosyncCmapCtlFactKey = "corosync-cmapctl"
)

type CorosyncCmapctlGatherer struct {
	executor CommandExecutor
}

func NewCorosyncCmapctlGatherer(executor CommandExecutor) *CorosyncCmapctlGatherer {
	return &CorosyncCmapctlGatherer{
		executor: executor,
	}
}

func (s *CorosyncCmapctlGatherer) Gather(factsRequests []FactRequest) ([]Fact, error) {
	facts := []Fact{}
	log.Infof("Starting %s facts gathering process", CorosyncCmapCtlFactKey)

	corosyncCmapctl, err := s.executor.Exec(
		"corosync-cmapctl", "-b")
	if err != nil {
		return facts, err
	}

	corosyncCmapctlMap := utils.FindMatches(`(?m)^(\S*)\s\(\S*\)\s=\s(.*)$`, corosyncCmapctl)

	for _, factReq := range factsRequests {
		if value, ok := corosyncCmapctlMap[factReq.Argument]; ok {
			fact := NewFactWithRequest(factReq, fmt.Sprint(value))
			facts = append(facts, fact)
		} else {
			log.Warnf("%s gatherer: requested fact %s not found", CorosyncCmapCtlFactKey, factReq.Argument)
		}
	}

	log.Infof("Requested %s facts gathered", CorosyncCmapCtlFactKey)
	return facts, nil
}
