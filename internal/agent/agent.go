package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"

	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/internal/factsengine"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

const machineIDPath = "/etc/machine-id"

var (
	fileSystem      = afero.NewOsFs()                                               //nolint
	trentoNamespace = uuid.Must(uuid.Parse("fb92284e-aa5e-47f6-a883-bf9469e7a0dc")) //nolint
)

type Agent struct {
	agentID         string
	config          *Config
	collectorClient collector.Client
	discoveries     []discovery.Discovery
}

type Config struct {
	InstanceName       string
	DiscoveriesConfig  *discovery.DiscoveriesConfig
	FactsEngineEnabled bool
	FactsServiceURL    string
	PluginsFolder      string
}

// NewAgent returns a new instance of Agent with the given configuration
func NewAgent(config *Config) (*Agent, error) {
	agentID, err := GetAgentID()
	if err != nil {
		return nil, errors.Wrap(err, "could not get the agent ID")
	}

	config.DiscoveriesConfig.CollectorConfig.AgentID = agentID

	collectorClient := collector.NewCollectorClient(config.DiscoveriesConfig.CollectorConfig)

	discoveries := []discovery.Discovery{
		discovery.NewClusterDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewSAPSystemsDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewCloudDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewSubscriptionDiscovery(collectorClient, config.InstanceName, *config.DiscoveriesConfig),
		discovery.NewHostDiscovery(collectorClient, config.InstanceName, *config.DiscoveriesConfig),
	}

	agent := &Agent{
		agentID:         agentID,
		config:          config,
		collectorClient: collectorClient,
		discoveries:     discoveries,
	}
	return agent, nil
}

func GetAgentID() (string, error) {
	machineIDBytes, err := afero.ReadFile(fileSystem, machineIDPath)
	if err != nil {
		return "", err
	}

	machineID := strings.TrimSpace(string(machineIDBytes))
	agentID := uuid.NewSHA1(trentoNamespace, []byte(machineID))

	return agentID.String(), nil
}

// Start the Agent. This will start the discovery ticker and the heartbeat ticker
func (a *Agent) Start(ctx context.Context) error {
	g, groupCtx := errgroup.WithContext(ctx)

	for _, d := range a.discoveries {
		dLoop := d
		g.Go(func() error {
			log.Infof("Starting %s loop...", dLoop.GetID())
			a.startDiscoverTicker(groupCtx, dLoop)
			log.Infof("%s discover loop stopped.", dLoop.GetID())
			return nil
		})
	}

	g.Go(func() error {
		log.Info("Starting heartbeat loop...")
		a.startHeartbeatTicker(groupCtx)
		log.Info("heartbeat loop stopped.")
		return nil
	})

	if a.config.FactsEngineEnabled {

		gathererRegistry := gatherers.NewRegistry(gatherers.StandardGatherers())

		log.Info("loading plugins")

		pluginLoaders := gatherers.PluginLoaders{
			"rpc": &gatherers.RPCPluginLoader{},
		}

		gatherersFromPlugins, err := gatherers.GetGatherersFromPlugins(
			pluginLoaders,
			a.config.PluginsFolder,
		)
		if err != nil {
			log.Fatalf("Error loading gatherers from plugins: %s", err)
		}

		gathererRegistry.AddGatherers(gatherersFromPlugins)

		c := factsengine.NewFactsEngine(a.agentID, a.config.FactsServiceURL, *gathererRegistry)

		g.Go(func() error {
			log.Info("Starting fact gathering service...")
			if err := c.Subscribe(); err != nil {
				return err
			}

			if err := c.Listen(groupCtx); err != nil {
				return err
			}

			log.Info("fact gathering stopped.")
			return nil
		})
	}

	return g.Wait()
}

func (a *Agent) Stop(ctxCancel context.CancelFunc) {
	ctxCancel()
}

// Start a Ticker loop that will iterate over the hardcoded list of Discovery backends and execute them.
func (a *Agent) startDiscoverTicker(ctx context.Context, d discovery.Discovery) {

	tick := func() {
		result, err := d.Discover()
		if err != nil {
			result = fmt.Sprintf("Error while running discovery '%s': %s", d.GetID(), err)
			log.Errorln(result)
		}
		log.Infof("%s discovery tick output: %s", d.GetID(), result)
	}
	repeat(ctx, d.GetID(), tick, d.GetInterval())

}

func (a *Agent) startHeartbeatTicker(ctx context.Context) {
	tick := func() {
		err := a.collectorClient.Heartbeat()
		if err != nil {
			log.Errorf("Error while sending the heartbeat to the server: %s", err)
		}
	}

	repeat(ctx, "agent.heartbeat", tick, HeartbeatInterval)
}

// Repeat executes a function at a given interval.
// the first tick runs immediately
func repeat(ctx context.Context, operation string, tick func(), interval time.Duration) {
	tick()

	ticker := time.NewTicker(interval)
	msg := fmt.Sprintf("Next execution for operation %s in %s", operation, interval)
	log.Debugf(msg)

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tick()
			log.Debugf(msg)
		case <-ctx.Done():
			return
		}
	}
}
