package discovery

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/internal/discovery/collector"
)

const ClusterDiscoveryID string = "ha_cluster_discovery"
const ClusterDiscoveryMinPeriod time.Duration = 1 * time.Second

// This Discover handles any Pacemaker Cluster type
type ClusterDiscovery struct {
	id              string
	collectorClient collector.Client
	interval        time.Duration
}

func NewClusterDiscovery(collectorClient collector.Client, config DiscoveriesConfig) Discovery {
	return ClusterDiscovery{
		collectorClient: collectorClient,
		id:              ClusterDiscoveryID,
		interval:        config.DiscoveriesPeriodsConfig.Cluster,
	}
}

func (c ClusterDiscovery) GetID() string {
	return c.id
}

func (c ClusterDiscovery) GetInterval() time.Duration {
	return c.interval
}

// Execute one iteration of a discovery and publish the results to the collector
func (c ClusterDiscovery) Discover(ctx context.Context) (string, error) {
	cluster, err := cluster.NewCluster()

	if err != nil {
		slog.Debug("Error creating the cluster data object", "error", err)
	}

	err = c.collectorClient.Publish(ctx, c.id, cluster)
	if err != nil {
		slog.Debug("Error while sending cluster discovery to data collector", "error", err)
		return "", err
	}

	// If no cluster is found, the discovery payload is sent anyway.
	// This is used by the delta deregstration mechanism to remove a node from a cluster,
	// when the node is no longer part of a cluster.
	if cluster == nil {
		return "No HA cluster discovered on this host", nil
	}

	return fmt.Sprintf("Cluster with name: %s successfully discovered", cluster.Name), nil
}
