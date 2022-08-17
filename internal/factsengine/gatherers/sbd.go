package gatherers

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/cluster"
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

func (g *SBDGatherer) Gather(factsRequests []FactRequest) ([]Fact, error) {
	gatheredFacts := []Fact{}
	log.Infof("Starting SBD Facts gathering")

	conf, err := godotenv.Read(g.configFile)

	if err != nil {
		log.Errorf("Unable to parse SBD configuration file: %s", g.configFile)
		return gatheredFacts, err
	}

	for _, requestedFact := range factsRequests {
		var fact Fact
		if value, found := conf[requestedFact.Argument]; found {
			fact = NewFactWithRequest(requestedFact, value)
		} else {
			log.Infof("Requested SBD configuration '%s' was not found in the config file", requestedFact.Argument)
			fact = NewFactWithRequest(requestedFact, UndefinedSBDConfig)
		}

		gatheredFacts = append(gatheredFacts, fact)
	}

	return gatheredFacts, nil
}
