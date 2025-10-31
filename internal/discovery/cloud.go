package discovery

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/pkg/utils"
)

const CloudDiscoveryID string = "cloud_discovery"
const CloudDiscoveryMinPeriod time.Duration = 1 * time.Second

type CloudDiscovery struct {
}

func NewCloudDiscovery() Discoverer[*cloud.Instance] {
	return CloudDiscovery{}
}

func (d CloudDiscovery) Discover(ctx context.Context) (*cloud.Instance, string, error) {
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: 30 * time.Second}
	cloudData, err := cloud.NewCloudInstance(ctx, utils.Executor{}, client)
	if err != nil {
		return nil, "", err
	}

	if cloudData.Provider == "" {
		return cloudData, "No cloud provider discovered on this host", nil
	}

	return cloudData, fmt.Sprintf("Cloud provider %s discovered", cloudData.Provider), nil
}
