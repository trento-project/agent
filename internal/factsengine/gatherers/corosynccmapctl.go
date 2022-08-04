package gatherers

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/utils"
)

const (
	CorosyncCmapCtlFactKey = "corosync-cmapctl"
)

type CorosyncCmapctlGatherer struct {
	executor CommandExecutor
}

func NewCorosyncCmapctlGatherer() *CorosyncCmapctlGatherer {
	return &CorosyncCmapctlGatherer{
		executor: Executor{},
	}
}

func (s *CorosyncCmapctlGatherer) Gather(factsRequests []FactRequest) ([]Fact, error) {
	facts := []Fact{}
	log.Infof("Starting corosync-cmapctl facts gathering process")

	corosyncCmapctl, err := s.executor.Exec( // nolint:gosec
		"corosync-cmapctl", "-b")
	if err != nil {
		return facts, err
	}

	corosyncCmapctlMap := utils.FindMatches(`(?m)^(\S*)\s\(\S*\)\s=\s(.*)$`, corosyncCmapctl)

	for _, factReq := range factsRequests {
		if value, ok := corosyncCmapctlMap[factReq.Argument]; ok {
			facts = append(facts, Fact{
				Name:  factReq.Name,
				Value: value,
			})
		}
	}

	log.Infof("Requested corosync-cmapctl facts gathered")
	return facts, nil
}
