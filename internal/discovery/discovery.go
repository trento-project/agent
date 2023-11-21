package discovery

import (
	"context"
	"time"

	"github.com/trento-project/agent/internal/discovery/collector"
)

type DiscoveriesPeriodConfig struct {
	Cluster      time.Duration
	SAPSystem    time.Duration
	Cloud        time.Duration
	Host         time.Duration
	Subscription time.Duration
	Saptune      time.Duration
}

type DiscoveriesConfig struct {
	DiscoveriesPeriodsConfig *DiscoveriesPeriodConfig
	CollectorConfig          *collector.Config
}

type Discovery interface {
	// Returns an arbitrary unique string identifier of the discovery
	GetID() string
	// Execute the discovery mechanism
	Discover(ctx context.Context) (string, error)
	// Get interval
	GetInterval() time.Duration
}
