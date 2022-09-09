package discovery

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/internal/sapsystem"
)

const SAPDiscoveryID string = "sap_system_discovery"
const SAPDiscoveryMinPeriod time.Duration = 1 * time.Second

type SAPSystemsDiscovery struct {
	id              string
	collectorClient collector.Client
	interval        time.Duration
}

func NewSAPSystemsDiscovery(collectorClient collector.Client, config DiscoveriesConfig) Discovery {
	return SAPSystemsDiscovery{
		id:              SAPDiscoveryID,
		collectorClient: collectorClient,
		interval:        config.DiscoveriesPeriodsConfig.SAPSystem,
	}
}

func (d SAPSystemsDiscovery) GetID() string {
	return d.id
}

func (d SAPSystemsDiscovery) GetInterval() time.Duration {
	return d.interval
}

func (d SAPSystemsDiscovery) Discover() (string, error) {
	systems, err := sapsystem.NewSAPSystemsList()

	if err != nil {
		return "", err
	}

	err = d.collectorClient.Publish(d.id, systems)
	if err != nil {
		log.Debugf("Error while sending sapsystem discovery to data collector: %s", err)
		return "", err
	}

	sysNames := systems.GetSIDsString()
	if sysNames != "" {

		return fmt.Sprintf("SAP system(s) with ID: %s discovered", sysNames), nil
	}

	return "No SAP system discovered on this host", nil
}
