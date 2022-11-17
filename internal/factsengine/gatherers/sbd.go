package gatherers

import (
	"os"

	"github.com/hashicorp/go-envparse"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/cluster"
	"github.com/trento-project/agent/internal/factsengine/entities"
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

	SBDConfigDecodingError = entities.FactGatheringError{
		Type:    "sbd-config-decoding-error",
		Message: "error decoding configuration file",
	}

	SBDConfigValueNotFoundError = entities.FactGatheringError{
		Type:    "sbd-config-value-not-found",
		Message: "requested field value not found",
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

func (g *SBDGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting SBD config Facts gathering")

	conf, err := loadSbdConf(g.configFile)

	if err != nil {
		return nil, err
	}

	for _, requestedFact := range factsRequests {
		var fact entities.Fact
		if value, found := conf[requestedFact.Argument]; found {
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

func loadSbdConf(sbdConfigPath string) (map[string]string, error) {
	sbdConfigFile, err := os.Open(sbdConfigPath)
	if err != nil {
		return nil, SBDConfigFileError.Wrap(err.Error())
	}

	defer func() {
		err := sbdConfigFile.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	conf, err := envparse.Parse(sbdConfigFile)
	if err != nil {
		return nil, SBDConfigDecodingError.Wrap(err.Error())
	}

	return conf, nil
}
