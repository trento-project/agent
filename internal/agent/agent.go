// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"log/slog"

	"golang.org/x/sync/errgroup"

	"github.com/trento-project/agent/v3/internal/discovery"
	"github.com/trento-project/agent/v3/internal/discovery/collector"
	"github.com/trento-project/agent/v3/internal/factsengine"
	"github.com/trento-project/agent/v3/internal/factsengine/gatherers"
	"github.com/trento-project/agent/v3/internal/operations"
	"github.com/trento-project/agent/v3/internal/operations/operator"
)

type Agent struct {
	config          *Config
	collectorClient collector.Client
	discoveries     []discovery.Discovery
}

type Config struct {
	AgentID           string
	InstanceName      string
	DiscoveriesConfig *discovery.DiscoveriesConfig
	FactsServiceURL   string
	PluginsFolder     string
	PrometheusConfig  *discovery.PrometheusConfig
	HeartbeatInterval time.Duration
}

// NewAgent returns a new instance of Agent with the given configuration.
func NewAgent(config *Config) (*Agent, error) {
	agentClient := http.Client{Timeout: 30 * time.Second}
	collectorClient := collector.NewCollectorClient(config.DiscoveriesConfig.CollectorConfig, &agentClient)

	discoveries := []discovery.Discovery{
		discovery.NewClusterDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewSAPSystemsDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewCloudDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewSubscriptionDiscovery(collectorClient, config.InstanceName, *config.DiscoveriesConfig),
		discovery.NewHostDiscovery(collectorClient,
			config.InstanceName,
			*config.PrometheusConfig,
			*config.DiscoveriesConfig),
		discovery.NewSaptuneDiscovery(collectorClient, *config.DiscoveriesConfig),
	}

	agent := &Agent{
		config:          config,
		collectorClient: collectorClient,
		discoveries:     discoveries,
	}

	return agent, nil
}

// Start the Agent. This will start the discovery ticker and the heartbeat ticker.
func (a *Agent) Start(ctx context.Context) error {
	gathererRegistry := gatherers.NewRegistry(
		gatherers.StandardGatherers(
			gatherers.Config{AgentID: a.config.AgentID},
		),
	)

	c := factsengine.NewFactsEngine(a.config.AgentID, a.config.FactsServiceURL, *gathererRegistry)

	slog.Info("Starting fact gathering service...")

	err := c.Subscribe()
	if err != nil {
		return err
	}

	operatorsRegistry := operator.StandardRegistry()

	op := operations.NewOperationsEngine(a.config.AgentID, a.config.FactsServiceURL, *operatorsRegistry)

	slog.Info("Starting operations service...")

	err = op.Subscribe()
	if err != nil {
		return err
	}

	g, groupCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := c.Listen(groupCtx)
		if err != nil {
			return err
		}

		slog.Info("fact gathering stopped.")

		return nil
	})

	g.Go(func() error {
		err := op.Listen(groupCtx)
		if err != nil {
			return err
		}

		slog.Info("operators execution stopped.")

		return nil
	})

	g.Go(func() error {
		err := discovery.ListenRequests(
			groupCtx,
			a.config.AgentID,
			a.config.FactsServiceURL,
			a.discoveries,
		)
		if err != nil {
			return err
		}

		slog.Info("discovery requests listener stopped.")

		return nil
	})

	for _, d := range a.discoveries {
		dLoop := d

		g.Go(func() error {
			slog.Info("Starting loop", "id", dLoop.GetID())
			a.startDiscoverTicker(groupCtx, dLoop)
			slog.Info("discover loop stopped", "id", dLoop.GetID())

			return nil
		})
	}

	g.Go(func() error {
		slog.Info("Starting heartbeat loop...")
		a.startHeartbeatTicker(groupCtx)
		slog.Info("heartbeat loop stopped.")

		return nil
	})

	slog.Info("loading plugins")

	pluginLoaders := gatherers.PluginLoaders{
		"rpc": &gatherers.RPCPluginLoader{},
	}

	gatherersFromPlugins, err := gatherers.GetGatherersFromPlugins(
		pluginLoaders,
		a.config.PluginsFolder,
	)
	if err != nil {
		slog.Error("Error loading gatherers from plugins", "error", err)
		os.Exit(1)
	}

	gathererRegistry.AddGatherers(gatherersFromPlugins)

	return g.Wait()
}

func (a *Agent) Stop(ctxCancel context.CancelFunc) {
	ctxCancel()
}

// Start a Ticker loop that will iterate over the hardcoded list of Discovery backends and execute them.
func (a *Agent) startDiscoverTicker(ctx context.Context, d discovery.Discovery) {
	tick := func() {
		result, err := d.Discover(ctx)
		if err != nil {
			slog.Error("Error while running discovery", "discovery", d.GetID(), "error", err)
		}

		slog.Info("Discovery tick completed", "id", d.GetID(), "output", result)
	}
	repeat(ctx, d.GetID(), tick, d.GetInterval())
}

func (a *Agent) startHeartbeatTicker(ctx context.Context) {
	tick := func() {
		err := a.collectorClient.Heartbeat(ctx)
		if err != nil {
			slog.Error("Error while sending the heartbeat to the server", "error", err)
		}
	}

	repeat(ctx, "agent.heartbeat", tick, a.config.HeartbeatInterval)
}

// Repeat executes a function at a given interval.
// the first tick runs immediately.
func repeat(ctx context.Context, operation string, tick func(), interval time.Duration) {
	tick()

	ticker := time.NewTicker(interval)
	msg := fmt.Sprintf("Next execution for operation %s in %s", operation, interval)
	slog.Debug(msg)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tick()
			slog.Debug(msg)
		case <-ctx.Done():
			return
		}
	}
}
