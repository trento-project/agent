package gatherers

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SBDDumpGathererName = "sbd_dump"
)

type SBDDumpGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultSBDDumpGatherer() *SBDDumpGatherer {
	return NewSBDDumpGatherer(utils.Executor{})
}

func NewSBDDumpGatherer(executor utils.CommandExecutor) *SBDDumpGatherer {
	return &SBDDumpGatherer{
		executor: executor,
	}
}

func (gatherer *SBDDumpGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", SBDDumpGathererName)

	for _, factReq := range factsRequests {
		//FIXME: This is a workaround until we allow multiple arguments per fact request
		args := strings.Split(factReq.Argument, ":")
		SBDDump, err := gatherer.executor.Exec(
			"sbd", "dump", "-d", args[0], "dump")
		if err != nil {
			log.Errorf("Error getting sbd dump for device: %s", args[0])
			return facts, err
		}

		SBDDumpMap := utils.FindMatches(`(?m)^(\S+(?: \S+)*)\s*:\s(\S*)$`, SBDDump)
		key := strings.ReplaceAll(args[1], " ", "_")
		if value, ok := SBDDumpMap[key]; ok {
			fact := entities.NewFactGatheredWithRequest(factReq, entities.ParseStringToFactValue(fmt.Sprint(value)))
			facts = append(facts, fact)
		} else {
			log.Warnf("%s gatherer: requested fact %s not found", SBDDumpGathererName, factReq.Argument)
		}
	}

	log.Infof("Requested %s facts gathered", SBDDumpGathererName)
	return facts, nil
}
