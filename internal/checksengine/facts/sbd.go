package facts

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/cluster"
)

const (
	SBDFactKey    = "sbd_config"
	SBDConfigPath = "/etc/sysconfig/sbd"
)

type sbdConfigGatherer struct {
}

func NewSbdConfigGatherer() *sbdConfigGatherer {
	return &sbdConfigGatherer{}
}

func (s *sbdConfigGatherer) Gather(factsRequests []FactRequest) ([]*Fact, error) {
	var facts []*Fact
	log.Infof("Starting SBD configuration file facts gathering process")

	sbdConfig, err := cluster.GetSBDConfig(SBDConfigPath)
	if err != nil {
		return nil, err
	}

	for _, factReq := range factsRequests {
		value := sbdConfig[factReq.Name]
		fact := &Fact{
			Name:  SBDFactKey,
			Key:   factReq.Name,
			Value: value,
			Alias: factReq.Alias,
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested SBD configuration file facts gathered")
	return facts, nil
}
