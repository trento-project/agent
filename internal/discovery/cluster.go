package discovery

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/trento-project/agent/internal/core/cluster"
)

const ClusterDiscoveryID string = "ha_cluster_discovery"
const ClusterDiscoveryMinPeriod time.Duration = 1 * time.Second

// This Discover handles any Pacemaker Cluster type
type ClusterDiscovery struct{}

func NewClusterDiscovery() Discoverer[*cluster.Cluster] {
	return ClusterDiscovery{}
}

// Execute one iteration of a discovery and publish the results to the collector
func (c ClusterDiscovery) Discover(_ context.Context) (*cluster.Cluster, string, error) {
	cluster, err := cluster.NewCluster()

	if err != nil {
		slog.Debug("Error creating the cluster data object", "error", err)
	}

	// If no cluster is found, the discovery payload is sent anyway.
	// This is used by the delta deregstration mechanism to remove a node from a cluster,
	// when the node is no longer part of a cluster.
	if cluster == nil {
		return cluster, "No HA cluster discovered on this host", nil
	}

	return cluster, fmt.Sprintf("Cluster with name: %s successfully discovered", cluster.Name), nil
}
