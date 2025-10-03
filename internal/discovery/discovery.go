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

type Discovery[T any] interface {
	// Returns an arbitrary unique string identifier of the discovery
	GetID() string
	// Fetches data without publishing it
	Discover(ctx context.Context) (T, error)
	// Execute the discovery mechanism
	DiscoverAndPublish(ctx context.Context) (string, error)
	// Get interval
	GetInterval() time.Duration
}
