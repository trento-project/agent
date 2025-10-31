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

// Discoverer is a generic interface for discovery mechanisms
type Discoverer[T any] interface {
	// Execute the discovery mechanism
	// Returns the discovered data, a human-readable message and an error if something goes wrong
	Discover(ctx context.Context) (T, string, error)
}
