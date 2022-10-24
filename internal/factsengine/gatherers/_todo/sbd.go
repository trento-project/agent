package gatherers

import (
	"os"

	"github.com/hashicorp/go-envparse"
	"github.com/pkg/errors"

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

func (g *SBDGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting SBD Facts gathering")

	sbdConfigFile, err := os.Open(g.configFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not open SBD Config file")
	}

	defer func() {
		err := sbdConfigFile.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	conf, err := envparse.Parse(sbdConfigFile)
	if err != nil {
		return facts, errors.Wrapf(err, "Unable to parse SBD configuration file: %s", g.configFile)
	}

	for _, requestedFact := range factsRequests {
		var fact entities.Fact
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
