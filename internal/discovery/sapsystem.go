package discovery

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/internal/discovery/collector"
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

func (d SAPSystemsDiscovery) Discover(ctx context.Context) (string, error) {
	systems, err := sapsystem.NewDefaultSAPSystemsList(ctx)

	if err != nil {
		return "", err
	}

	err = d.collectorClient.Publish(ctx, d.id, systems)
	if err != nil {
		slog.Debug("Error while sending sapsystem discovery to data collector", "error", err)
		return "", err
	}

	sysNames := systems.GetSIDsString()
	if sysNames != "" {

		return fmt.Sprintf("SAP system(s) with ID: %s discovered", sysNames), nil
	}

	return "No SAP system discovered on this host", nil
}
