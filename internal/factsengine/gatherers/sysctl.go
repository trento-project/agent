package gatherers

import (
	"context"
	"strings"

	"log/slog"

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

func (s *SysctlGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	slog.Info("Starting facts gathering process", "gatherer", SysctlGathererName)

	output, err := s.executor.OutputContext(ctx, "/sbin/sysctl", "-a")
	if err != nil {
		return nil, SysctlCommandError.Wrap(err.Error())
	}

	sysctlMap := sysctlOutputToMap(output)
	for _, factReq := range factsRequests {
		var fact entities.Fact

		if len(factReq.Argument) == 0 {
			slog.Error(SysctlMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &SysctlMissingArgument)
		} else if value, err := sysctlMap.GetValue(factReq.Argument); err == nil {
			fact = entities.NewFactGatheredWithRequest(factReq, value)
		} else {
			gatheringError := SysctlValueNotFound.Wrap(factReq.Argument)
			slog.Error(gatheringError.Error())
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
		}

		facts = append(facts, fact)
	}

	slog.Info("Requested facts gathered", "gatherer", SysctlGathererName)
	return facts, nil
}

func sysctlOutputToMap(output []byte) *entities.FactValueMap {
	outputMap := &entities.FactValueMap{Value: make(map[string]entities.FactValue)}

	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(line) == 0 || len(parts) != 2 {
			slog.Error("Invalid sysctl output line", "line", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		cursor := outputMap
		pathComponents := strings.Split(key, ".")

		for i, component := range pathComponents {
			if i == len(pathComponents)-1 {
				cursor.Value[component] = entities.ParseStringToFactValue(value)
			} else if nestedMap, ok := cursor.Value[component].(*entities.FactValueMap); !ok {
				newMap := &entities.FactValueMap{Value: make(map[string]entities.FactValue)}
				cursor.Value[component] = newMap
				cursor = newMap
			} else {
				cursor = nestedMap
			}
		}
	}

	return outputMap
}
