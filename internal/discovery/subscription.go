package discovery

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/trento-project/agent/internal/core/subscription"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/pkg/utils"
)

const SubscriptionDiscoveryID string = "subscription_discovery"
const SubscriptionDiscoveryMinPeriod time.Duration = 20 * time.Second

type SubscriptionDiscovery struct {
	id              string
	collectorClient collector.Client
	host            string
	interval        time.Duration
}

func NewSubscriptionDiscovery(collectorClient collector.Client, hostname string, config DiscoveriesConfig) Discovery {
	return SubscriptionDiscovery{
		id:              SubscriptionDiscoveryID,
		host:            hostname,
		collectorClient: collectorClient,
		interval:        config.DiscoveriesPeriodsConfig.Subscription,
	}
}

func (d SubscriptionDiscovery) GetID() string {
	return d.id
}

func (d SubscriptionDiscovery) GetInterval() time.Duration {
	return d.interval
}

func (d SubscriptionDiscovery) Discover(ctx context.Context) (string, error) {
	subsData, err := subscription.NewSubscriptions(utils.Executor{})
	if err != nil {
		return "", err
	}

	err = d.collectorClient.Publish(ctx, d.id, subsData)
	if err != nil {
		slog.Debug("Error while sending subscription discovery to data collector", "error", err.Error())
		return "", err
	}

	return fmt.Sprintf("Subscription (%d entries) discovered", len(subsData)), nil
}
