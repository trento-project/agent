package gatherers

import (
	"encoding/json"
	"strings"

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
	SaptuneVersionUnsupported = entities.FactGatheringError{
		Type:    "saptune-version-not-supported",
		Message: "currently installed version of saptune is not supported",
	}

	SaptuneUnknownArgument = entities.FactGatheringError{
		Type:    "saptune-unknown-error",
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

type CachedFactValue struct {
	factValue    entities.FactValue
	factValueErr *entities.FactGatheringError
}

type SaptuneGatherer struct {
	executor         utils.CommandExecutor
	cachedFactValues map[string]CachedFactValue
}

func NewDefaultSaptuneGatherer() *SaptuneGatherer {
	return NewSaptuneGatherer(utils.Executor{})
}

func NewSaptuneGatherer(executor utils.CommandExecutor) *SaptuneGatherer {
	return &SaptuneGatherer{
		executor: executor,
	}
}

func parseJSONToFactValue(jsonStr json.RawMessage) (entities.FactValue, error) {
	// Unmarshal the JSON into an interface{} type.
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return nil, err
	}

	// Convert the parsed jsonData into a FactValue using NewFactValue.
	return entities.NewFactValue(jsonData)
}

func (s *SaptuneGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	s.cachedFactValues = make(map[string]CachedFactValue)

	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", SaptuneGathererName)
	saptuneRetriever, _ := saptune.NewSaptune(s.executor)
	for _, factReq := range factsRequests {
		var fact entities.Fact

		internalArguments, ok := whitelistedArguments[factReq.Argument]

		switch {
		case !saptuneRetriever.IsJSONSupported:
			log.Error(SaptuneVersionUnsupported.Message)
			fact = entities.NewFactGatheredWithError(factReq, &SaptuneVersionUnsupported)

		case len(internalArguments) > 0 && !ok:
			gatheringError := SaptuneUnknownArgument.Wrap(factReq.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(factReq, gatheringError)

		case len(internalArguments) == 0:
			log.Error(SaptuneMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(factReq, &SaptuneMissingArgument)

		default:
			factValue, err := handleArgument(&saptuneRetriever, internalArguments, s.cachedFactValues)
			if err != nil {
				fact = entities.NewFactGatheredWithError(factReq, err)
			} else {
				fact = entities.NewFactGatheredWithRequest(factReq, factValue)
			}
		}
		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SaptuneGathererName)
	return facts, nil
}

func handleArgument(
	saptuneRetriever *saptune.Saptune,
	arguments []string,
	cachedFactValues map[string]CachedFactValue,
) (entities.FactValue, *entities.FactGatheringError) {
	cacheKey := strings.Join(arguments, "-")
	if item, found := cachedFactValues[cacheKey]; found {
		log.Info("Using cached fact value")
		return item.factValue, item.factValueErr
	}

	saptuneOutput, commandError := saptuneRetriever.RunCommandJSON(arguments...)
	if commandError != nil {
		gatheringError := SaptuneCommandError.Wrap(commandError.Error())
		log.Error(gatheringError)
		updateCachedFactValue(nil, gatheringError, cacheKey, cachedFactValues)
		return nil, gatheringError
	}

	fv, err := parseJSONToFactValue(saptuneOutput)
	if err != nil {
		gatheringError := SaptuneCommandError.Wrap(err.Error())
		log.Error(gatheringError)
		updateCachedFactValue(nil, gatheringError, cacheKey, cachedFactValues)
		return nil, gatheringError
	}

	updateCachedFactValue(fv, nil, cacheKey, cachedFactValues)
	return fv, nil
}

func updateCachedFactValue(factValue entities.FactValue, factValueErr *entities.FactGatheringError, key string,
	cachedFactValues map[string]CachedFactValue) {
	log.Info("Updating cached fact value")
	cachedFactValues[key] = CachedFactValue{
		factValue:    factValue,
		factValueErr: factValueErr,
	}
}
