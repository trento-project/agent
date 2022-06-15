package facts

import (
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/cluster"
)

const (
	SBDFactKey    = "sbd_config"
	SBDConfigPath = "/etc/sysconfig/sbd"
)

func GatherSbdConfigFacts(keys []string) ([]*Fact, error) {
	var facts []*Fact
	log.Infof("Starting SBD configuration file facts gathering process")

	sbdConfig, err := cluster.GetSBDConfig(SBDConfigPath)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		value := sbdConfig[key]
		fact := &Fact{
			Name:  SBDFactKey,
			Key:   key,
			Value: value,
		}

		facts = append(facts, fact)
	}

	log.Infof("Requested SBD configuration file facts gathered")
	return facts, nil
}
