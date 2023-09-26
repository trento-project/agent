package gatherers

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/saptune"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
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

func (s *SaptuneGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	cachedFacts := make(map[string]entities.Fact)

	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", SaptuneGathererName)
	saptuneRetriever, err := saptune.NewSaptune(s.executor)

	for _, factReq := range factsRequests {
		var fact entities.Fact

		internalArguments, ok := whitelistedArguments[factReq.Argument]
		cachedFact, cacheHit := cachedFacts[factReq.Argument]

		switch {
		case err != nil:
			log.Error(err)
			fact = entities.NewFactGatheredWithError(factReq, &SaptuneNotInstalled)

		case !saptuneRetriever.IsJSONSupported:
			log.Error(SaptuneVersionUnsupported.Message)
			fact = entities.NewFactGatheredWithError(factReq, &SaptuneVersionUnsupported)

		case len(factReq.Argument) == 0:
			log.Error(SaptuneMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &SaptuneMissingArgument)

		case !ok:
			gatheringError := SaptuneArgumentUnsupported.Wrap(factReq.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)

		case cacheHit:
			fact = entities.Fact{
				Name:    factReq.Name,
				CheckID: factReq.CheckID,
				Value:   cachedFact.Value,
				Error:   cachedFact.Error,
			}

		default:
			factValue, err := runCommand(&saptuneRetriever, internalArguments)
			if err != nil {
				gatheringError := SaptuneCommandError.Wrap(err.Error())
				log.Error(gatheringError)
				fact = entities.NewFactGatheredWithError(factReq, gatheringError)
			} else {
				fact = entities.NewFactGatheredWithRequest(factReq, factValue)
			}
			cachedFacts[factReq.Argument] = fact
		}
		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SaptuneGathererName)
	return facts, nil
}

func runCommand(saptuneRetriever *saptune.Saptune, arguments []string) (entities.FactValue, error) {
	saptuneOutput, commandError := saptuneRetriever.RunCommandJSON(arguments...)
	if commandError != nil {
		return nil, commandError
	}

	log.Error(string(saptuneOutput))

	var jsonData interface{}
	if err := json.Unmarshal(saptuneOutput, &jsonData); err != nil {
		return nil, err
	}

	log.Error(jsonData)

	return entities.NewFactValue(jsonData, entities.WithSnakeCaseKeys())
}
