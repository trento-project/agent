package gatherers

import (
	"fmt"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/cluster"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

const (
	SBDConfigGathererName = "sbd_config"
)

var (
	SBDConfigParsingError = entities.FactGatheringError{ // nolint
		Type:    "sbd-config-parsing-error",
		Message: "error parsing SBD configuration file",
	}

	SBDValueNotFoundError = entities.FactGatheringError{ // nolint
		Type:    "sbd-config-value-not-found",
		Message: "requested field value not found",
	}
)

type SBDGatherer struct {
	configFile string
}

func NewSBDGathererWithDefaultConfig() *SBDGatherer {
	return NewSBDGatherer(cluster.SBDConfigPath)
}

func NewSBDGatherer(configFile string) *SBDGatherer {
	return &SBDGatherer{
		configFile,
	}
}

func (g *SBDGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.FactsGatheredItem, error) {
	facts := []entities.FactsGatheredItem{}
	log.Infof("Starting SBD Facts gathering")

	conf, err := godotenv.Read(g.configFile)

	if err != nil {
		gatheringError := SBDConfigParsingError.Wrap(err.Error())
		log.Errorf(gatheringError.Error())
		return entities.NewFactsGatheredListWithError(factsRequests, &gatheringError), nil
	}

	for _, requestedFact := range factsRequests {
		var fact entities.FactsGatheredItem

		if value, found := conf[requestedFact.Argument]; found {
			fact = entities.NewFactGatheredWithRequest(requestedFact, value)
		} else {
			gatheringError := SBDValueNotFoundError.Wrap(fmt.Sprintf("requested value %s not found", requestedFact.Argument))
			log.Errorf(gatheringError.Error())
			fact = entities.NewFactGatheredWithError(requestedFact, &gatheringError)
		}

		facts = append(facts, fact)
	}

	return facts, nil
}
