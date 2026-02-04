package gatherers

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/trento-project/agent/internal/core/saptune"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
	"golang.org/x/mod/semver"
)

const (
	SaptuneGathererName = "saptune"
)

// nolint:gochecknoglobals
var whitelistedArguments = map[string][]string{
	"status":          {"status", "--non-compliance-check"},
	"solution-verify": {"solution", "verify"},
	"solution-list":   {"solution", "list"},
	"note-verify":     {"note", "verify"},
	"note-list":       {"note", "list"},
	"check":           {"check"},
}

// Map to store supported saptune versions for specific commands.
// Arguments not present in this map use the global supported version
// nolint:gochecknoglobals
var argumentSupportedVersions = map[string]string{
	"check": "3.2.0",
}

// nolint:gochecknoglobals
var (
	SaptuneNotInstalled = entities.FactGatheringError{
		Type:    "saptune-not-installed",
		Message: "saptune is not installed",
	}

	SaptuneVersionUnsupported = entities.FactGatheringError{
		Type:    "saptune-version-not-supported",
		Message: "currently installed version of saptune is not supported",
	}

	SaptuneArgumentUnsupported = entities.FactGatheringError{
		Type:    "saptune-unsupported-argument",
		Message: "the requested argument is not currently supported",
	}

	SaptuneMissingArgument = entities.FactGatheringError{
		Type:    "saptune-missing-argument",
		Message: "missing required argument",
	}

	SaptuneCommandError = entities.FactGatheringError{
		Type:    "saptune-cmd-error",
		Message: "error executing saptune command",
	}
)

type SaptuneGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultSaptuneGatherer() *SaptuneGatherer {
	return NewSaptuneGatherer(utils.Executor{})
}

func NewSaptuneGatherer(executor utils.CommandExecutor) *SaptuneGatherer {
	return &SaptuneGatherer{
		executor: executor,
	}
}

func (s *SaptuneGatherer) Gather(_ context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	cachedFacts := make(map[string]entities.Fact)

	facts := []entities.Fact{}
	slog.Info("Starting facts gathering process", "gatherer", SaptuneGathererName)
	saptuneRetriever, err := saptune.NewSaptune(s.executor)
	if err != nil {
		return nil, SaptuneNotInstalled.Wrap(err.Error())
	}

	if !saptuneRetriever.IsJSONSupported {
		return nil, &SaptuneVersionUnsupported
	}

	for _, factReq := range factsRequests {
		var fact entities.Fact

		internalArguments, ok := whitelistedArguments[factReq.Argument]
		cachedFact, cacheHit := cachedFacts[factReq.Argument]

		switch {
		case len(factReq.Argument) == 0:
			slog.Error(SaptuneMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &SaptuneMissingArgument)

		case !ok:
			gatheringError := SaptuneArgumentUnsupported.Wrap(factReq.Argument)
			slog.Error(gatheringError.Error())
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)

		case cacheHit:
			fact = entities.Fact{
				Name:    factReq.Name,
				CheckID: factReq.CheckID,
				Value:   cachedFact.Value,
				Error:   cachedFact.Error,
			}

		case !isArgumentSupported(factReq.Argument, saptuneRetriever.Version):
			gatheringError := SaptuneVersionUnsupported.Wrap(factReq.Argument +
				" argument is not supported for saptune versions older than " + argumentSupportedVersions[factReq.Argument])
			slog.Error(gatheringError.Error())
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)

		default:
			factValue, err := runCommand(&saptuneRetriever, internalArguments)
			if err != nil {
				gatheringError := SaptuneCommandError.Wrap(err.Error())
				slog.Error(gatheringError.Error())
				fact = entities.NewFactGatheredWithError(factReq, gatheringError)
			} else {
				fact = entities.NewFactGatheredWithRequest(factReq, factValue)
			}
			cachedFacts[factReq.Argument] = fact
		}
		facts = append(facts, fact)
	}

	slog.Info("Requested facts gathered", "gatherer", SaptuneGathererName)
	return facts, nil
}

func isArgumentSupported(argument, saptuneVersion string) bool {
	supportedVersion, shouldCompare := argumentSupportedVersions[argument]
	if !shouldCompare {
		return true
	}

	return semver.Compare("v"+saptuneVersion, "v"+supportedVersion) >= 0
}

func runCommand(saptuneRetriever *saptune.Saptune, arguments []string) (entities.FactValue, error) {
	saptuneOutput, commandError := saptuneRetriever.RunCommandJSON(arguments...)
	if commandError != nil {
		return nil, commandError
	}

	var jsonData interface{}
	if err := json.Unmarshal(saptuneOutput, &jsonData); err != nil {
		return nil, err
	}

	return entities.NewFactValue(jsonData, entities.WithSnakeCaseKeys())
}
