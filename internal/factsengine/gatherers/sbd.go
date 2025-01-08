package gatherers

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	SBDConfigGathererName = "sbd_config"
)

// nolint:gochecknoglobals
var (
	SBDConfigFileError = entities.FactGatheringError{
		Type:    "sbd-config-file-error",
		Message: "error reading sbd configuration file",
	}

	SBDConfigValueNotFoundError = entities.FactGatheringError{
		Type:    "sbd-config-value-not-found",
		Message: "requested field value not found",
	}

	SBDConfigMissingArgument = entities.FactGatheringError{
		Type:    "sbd-config-missing-argument",
		Message: "missing required argument",
	}
)

type SBDGatherer struct {
	configFile string
}

func NewDefaultSBDGatherer() *SBDGatherer {
	return NewSBDGatherer(cluster.SBDConfigPath)
}

func NewSBDGatherer(configFile string) *SBDGatherer {
	return &SBDGatherer{
		configFile,
	}
}

func (g *SBDGatherer) Gather(_ context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting SBD config Facts gathering")

	conf, err := cluster.LoadSbdConfig(g.configFile)

	if err != nil {
		return nil, SBDConfigFileError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		var fact entities.Fact

		if len(requestedFact.Argument) == 0 {
			log.Error(SBDConfigMissingArgument.Message)
			fact = entities.NewFactGatheredWithError(requestedFact, &SBDConfigMissingArgument)
		} else if value, found := conf[requestedFact.Argument]; found {
			fact = entities.NewFactGatheredWithRequest(requestedFact, entities.ParseStringToFactValue(value))
		} else {
			gatheringError := SBDConfigValueNotFoundError.Wrap(requestedFact.Argument)
			log.Error(gatheringError)
			fact = entities.NewFactGatheredWithError(requestedFact, gatheringError)
		}

		facts = append(facts, fact)
	}

	return facts, nil
}
