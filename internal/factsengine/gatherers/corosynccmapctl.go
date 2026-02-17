package gatherers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	CorosyncCmapCtlGathererName = "corosync-cmapctl"
)

// nolint:gochecknoglobals
var (
	CorosyncCmapCtlValueNotFound = entities.FactGatheringError{
		Type:    "corosync-cmapctl-value-not-found",
		Message: "requested value not found in corosync-cmapctl output",
	}

	CorosyncCmapCtlCommandError = entities.FactGatheringError{
		Type:    "corosync-cmapctl-command-error",
		Message: "error while executing corosynccmap-ctl",
	}

	CorosyncCmapCtlMissingArgument = entities.FactGatheringError{
		Type:    "corosync-cmapctl-missing-argument",
		Message: "missing required argument",
	}
)

type CorosyncCmapctlGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultCorosyncCmapctlGatherer() *CorosyncCmapctlGatherer {
	return NewCorosyncCmapctlGatherer(utils.Executor{})
}

func NewCorosyncCmapctlGatherer(executor utils.CommandExecutor) *CorosyncCmapctlGatherer {
	return &CorosyncCmapctlGatherer{
		executor: executor,
	}
}

func corosyncCmapctlOutputToMap(corosyncCmapctlOutput string) *entities.FactValueMap {
	outputMap := &entities.FactValueMap{Value: make(map[string]entities.FactValue)}
	var cursor *entities.FactValueMap

	for _, line := range strings.Split(corosyncCmapctlOutput, "\n") {
		if len(line) == 0 {
			continue
		}

		cursor = outputMap

		value := strings.Split(line, "= ")[1]

		pathAsString := strings.Split(line, " (")[0]
		path := strings.Split(pathAsString, ".")

		for i, key := range path {
			currentMap := cursor

			if i == len(path)-1 {
				currentMap.Value[key] = entities.ParseStringToFactValue(value)

				break
			}

			if _, found := currentMap.Value[key]; !found {
				currentMap.Value[key] = &entities.FactValueMap{Value: make(map[string]entities.FactValue)}
			}

			cursor = currentMap.Value[key].(*entities.FactValueMap) //nolint:forcetypeassert
		}
	}

	return outputMap
}

func (s *CorosyncCmapctlGatherer) Gather(
	ctx context.Context,
	factsRequests []entities.FactRequest,
) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	slog.Info("Starting facts gathering process", "gatherer", CorosyncCmapCtlGathererName)

	corosyncCmapctl, err := s.executor.OutputContext(ctx,
		"corosync-cmapctl", "-b")
	if err != nil {
		return nil, CorosyncCmapCtlCommandError.Wrap(err.Error())
	}

	corosyncCmapctlMap := corosyncCmapctlOutputToMap(string(corosyncCmapctl))

	for _, factReq := range factsRequests {
		var fact entities.Fact

		if len(factReq.Argument) == 0 {
			slog.Error(CorosyncCmapCtlMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &CorosyncCmapCtlMissingArgument)
		} else if value, err := corosyncCmapctlMap.GetValue(factReq.Argument); err == nil {
			fact = entities.NewFactGatheredWithRequest(factReq, value)
		} else {
			gatheringError := CorosyncCmapCtlValueNotFound.Wrap(factReq.Argument)
			slog.Error(gatheringError.Error())
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)
		}

		facts = append(facts, fact)
	}

	slog.Info("Requested facts gathered", "gatherer", CorosyncCmapCtlGathererName)
	return facts, nil
}
