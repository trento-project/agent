package discovery

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/pkg/utils"
)

const CloudDiscoveryID string = "cloud_discovery"
const CloudDiscoveryMinPeriod time.Duration = 1 * time.Second

type CloudDiscovery struct {
	id              string
	collectorClient collector.Client
	interval        time.Duration
}

func NewCloudDiscovery(collectorClient collector.Client, config DiscoveriesConfig) Discovery {
	return CloudDiscovery{
		collectorClient: collectorClient,
		id:              CloudDiscoveryID,
		interval:        config.DiscoveriesPeriodsConfig.Cloud,
	}
}

func (d CloudDiscovery) GetID() string {
	return d.id
}

func (d CloudDiscovery) GetInterval() time.Duration {
	return d.interval
}

func (d CloudDiscovery) Discover(ctx context.Context) (string, error) {
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: 30 * time.Second}
	cloudData, err := cloud.NewCloudInstance(utils.Executor{}, client)
	if err != nil {
		return "", err
	}

	err = d.collectorClient.Publish(d.id, cloudData)
	if err != nil {
		log.Debugf("Error while sending cloud discovery to data collector: %s", err)
		return "", err
	}

	if cloudData.Provider == "" {
		return "No cloud provider discovered on this host", nil
	}

	return fmt.Sprintf("Cloud provider %s discovered", cloudData.Provider), nil
}
