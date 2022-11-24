package discovery

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
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

func (d SubscriptionDiscovery) Discover() (string, error) {
	subsData, err := subscription.NewSubscriptions(utils.Executor{})
	if err != nil {
		return "", err
	}

	err = d.collectorClient.Publish(d.id, subsData)
	if err != nil {
		log.Debugf("Error while sending subscription discovery to data collector: %s", err)
		return "", err
	}

	return fmt.Sprintf("Subscription (%d entries) discovered", len(subsData)), nil
}
