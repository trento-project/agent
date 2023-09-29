package gatherers

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SysctlGathererName = "sysctl"
)

// nolint:gochecknoglobals
var (
	SysctlValueNotFound = entities.FactGatheringError{
		Type:    "sysctl-value-not-found",
		Message: "requested value not found in sysctl output",
	}

	SysctlCommandError = entities.FactGatheringError{
		Type:    "sysctl-cmd-error",
		Message: "error executing sysctl command",
	}

	SysctlMissingArgument = entities.FactGatheringError{
		Type:    "sysctl-missing-argument",
		Message: "missing required argument",
	}
)

type SysctlGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultSysctlGatherer() *SysctlGatherer {
	return NewSysctlGatherer(utils.Executor{})
}

func NewSysctlGatherer(executor utils.CommandExecutor) *SysctlGatherer {
	return &SysctlGatherer{
		executor: executor,
	}
}

func (s *SysctlGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting sysctl facts gathering process")

	output, err := s.executor.Exec("sysctl", "-a")
	if err != nil {
		return nil, SysctlCommandError.Wrap(err.Error())
	}

	sysctlMap := sysctlOutputToMap(output)
	for _, factReq := range factsRequests {
		var fact entities.Fact

		if len(factReq.Argument) == 0 {
			log.Error(SysctlMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &SysctlMissingArgument)
		} else if value, err := sysctlMap.GetValue(factReq.Argument); err == nil {
			fact = entities.NewFactGatheredWithRequest(factReq, value)
		} else {
			gatheringError := SysctlValueNotFound.Wrap(factReq.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", CorosyncCmapCtlGathererName)
	return facts, nil
}

func sysctlOutputToMap(output []byte) *entities.FactValueMap {
	outputMap := &entities.FactValueMap{Value: make(map[string]entities.FactValue)}
	var cursor *entities.FactValueMap

	for _, line := range strings.Split(string(output), "\n") {
		if len(line) == 0 {
			continue
		}

		parts := strings.SplitN(strings.TrimSpace(line), " = ", 2)
		switch {
		case len(parts) == 1:
			parts[0] = strings.TrimSuffix(parts[0], " =")
			parts = append(parts, "")
		case len(parts) == 2:
		default:
			continue
		}

		value := parts[1]
		path := strings.Split(parts[0], ".")
		cursor = outputMap

		for i, key := range path {
			currentMap, ok := cursor.Value[key].(*entities.FactValueMap)
			if !ok {
				if i == len(path)-1 {
					cursor.Value[key] = entities.ParseStringToFactValue(value)
					break
				} else {
					currentMap = &entities.FactValueMap{Value: make(map[string]entities.FactValue)}
					cursor.Value[key] = currentMap
				}
			}

			cursor = currentMap
		}
	}

	return outputMap
}
