package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/trento-project/agent/internal/core/subscription"
	"github.com/trento-project/agent/pkg/utils"
)

const SubscriptionDiscoveryID string = "subscription_discovery"
const SubscriptionDiscoveryMinPeriod time.Duration = 20 * time.Second

type SubscriptionDiscovery struct {
	host string
}

func NewSubscriptionDiscovery(hostname string) Discoverer[subscription.Subscriptions] {
	return SubscriptionDiscovery{
		host: hostname,
	}
}

func (d SubscriptionDiscovery) Discover(_ context.Context) (subscription.Subscriptions, string, error) {
	subsData, err := subscription.NewSubscriptions(utils.Executor{})
	if err != nil {
		return nil, "", err
	}

	return subsData, fmt.Sprintf("Subscription (%d entries) discovered", len(subsData)), nil
}
