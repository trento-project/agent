package discovery

import (
	"context"

	"github.com/trento-project/agent/internal/discovery/collector"
)

type PublisherRegistry = map[string]Publisher

func DefaultPublisherRegistry(
	collectorClient collector.Client,
	discoveriesConfig *DiscoveriesConfig,
	instanceName string,
	prometheusTargets PrometheusTargets,
) PublisherRegistry {
	return PublisherRegistry{

		SAPDiscoveryID: NewDefaultPublisher(
			SAPDiscoveryID,
			discoveriesConfig.DiscoveriesPeriodsConfig.SAPSystem,
			collectorClient,
			func(ctx context.Context) (any, string, error) {
				return NewSAPSystemsDiscovery().Discover(ctx)
			},
		),

		ClusterDiscoveryID: NewDefaultPublisher(
			ClusterDiscoveryID,
			discoveriesConfig.DiscoveriesPeriodsConfig.Cluster,
			collectorClient,
			func(ctx context.Context) (any, string, error) {
				return NewClusterDiscovery().Discover(ctx)
			},
		),

		CloudDiscoveryID: NewDefaultPublisher(
			CloudDiscoveryID,
			discoveriesConfig.DiscoveriesPeriodsConfig.Cloud,
			collectorClient,
			func(ctx context.Context) (any, string, error) {
				return NewCloudDiscovery().Discover(ctx)
			},
		),

		SubscriptionDiscoveryID: NewDefaultPublisher(
			SubscriptionDiscoveryID,
			discoveriesConfig.DiscoveriesPeriodsConfig.Subscription,
			collectorClient,
			func(ctx context.Context) (any, string, error) {
				return NewSubscriptionDiscovery(instanceName).Discover(ctx)
			},
		),

		HostDiscoveryID: NewDefaultPublisher(
			HostDiscoveryID,
			discoveriesConfig.DiscoveriesPeriodsConfig.Host,
			collectorClient,
			func(ctx context.Context) (any, string, error) {
				return NewHostDiscovery(instanceName, prometheusTargets).Discover(ctx)
			},
		),

		SaptuneDiscoveryID: NewDefaultPublisher(
			SaptuneDiscoveryID,
			discoveriesConfig.DiscoveriesPeriodsConfig.Saptune,
			collectorClient,
			func(ctx context.Context) (any, string, error) {
				return NewSaptuneDiscovery().Discover(ctx)
			},
		),
	}
}
