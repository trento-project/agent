package gatherers

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/cluster"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

const (
	SBDConfigGathererName = "sbd_config"
	UndefinedSBDConfig    = "trento.checks.sbd.errors.undefined_configuration"
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
		log.Errorf("Unable to parse SBD configuration file: %s", g.configFile)
		return facts, err
	}

	for _, requestedFact := range factsRequests {
		var fact entities.FactsGatheredItem
		if value, found := conf[requestedFact.Argument]; found {
			fact = entities.NewFactGatheredWithRequest(requestedFact, value)
		} else {
			log.Infof("Requested SBD configuration '%s' was not found in the config file", requestedFact.Argument)
			fact = entities.NewFactGatheredWithRequest(requestedFact, UndefinedSBDConfig)
		}

		facts = append(facts, fact)
	}

	return facts, nil
}
