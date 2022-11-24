package discovery

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/pkg/discovery/collector"
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
func (c ClusterDiscovery) Discover() (string, error) {
	cluster, err := cluster.NewCluster()
	if err != nil {
		log.Debugf("Error creating the cluster data object: %s", err)
		return "No HA cluster discovered on this host", nil
	}

	err = c.collectorClient.Publish(c.id, cluster)
	if err != nil {
		log.Debugf("Error while sending cluster discovery to data collector: %s", err)
		return "", err
	}

	return fmt.Sprintf("Cluster with name: %s successfully discovered", cluster.Name), nil
}
