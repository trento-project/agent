package discovery

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/saptune"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/pkg/utils"
)

const SaptuneDiscoveryID string = "saptune_discovery"
const SaptuneDiscoveryMinPeriod time.Duration = 1 * time.Second

type SaptuneDiscovery struct {
	id              string
	collectorClient collector.Client
	host            string
	interval        time.Duration
}

func NewSaptuneDiscovery(collectorClient collector.Client, hostname string, config DiscoveriesConfig) Discovery {
	return SaptuneDiscovery{
		id:              SaptuneDiscoveryID,
		host:            hostname,
		collectorClient: collectorClient,
		interval:        config.DiscoveriesPeriodsConfig.Saptune,
	}
}

func (d SaptuneDiscovery) GetID() string {
	return d.id
}

func (d SaptuneDiscovery) GetInterval() time.Duration {
	return d.interval
}

func (d SaptuneDiscovery) Discover() (string, error) {
	saptuneData, err := saptune.NewSaptune(utils.Executor{})
	if err != nil {
		return "", err
	}

	err = d.collectorClient.Publish(d.id, saptuneData)
	if err != nil {
		log.Debugf("Error while sending saptune discovery to data collector: %s", err)
		return "", err
	}

	return fmt.Sprintf("Saptune data discovery completed"), nil
}
