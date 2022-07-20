package internal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/internal/factsengine"
)

const machineIdPath = "/etc/machine-id"

var (
	fileSystem      = afero.NewOsFs()
	trentoNamespace = uuid.Must(uuid.Parse("fb92284e-aa5e-47f6-a883-bf9469e7a0dc"))
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
	FactsServiceUrl    string
}

// NewAgent returns a new instance of Agent with the given configuration
func NewAgent(config *Config) (*Agent, error) {
	agentID, err := getAgentID()
	if err != nil {
		return nil, errors.Wrap(err, "could not get the agent ID")
	}

	config.DiscoveriesConfig.CollectorConfig.AgentID = agentID

	collectorClient := collector.NewCollectorClient(config.DiscoveriesConfig.CollectorConfig)

	discoveries := []discovery.Discovery{
		discovery.NewClusterDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewSAPSystemsDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewCloudDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewSubscriptionDiscovery(collectorClient, *config.DiscoveriesConfig),
		discovery.NewHostDiscovery(collectorClient, *config.DiscoveriesConfig),
	}

	agent := &Agent{
		agentID:         agentID,
		config:          config,
		collectorClient: collectorClient,
		discoveries:     discoveries,
	}
	return agent, nil
}

func getAgentID() (string, error) {
	machineIDBytes, err := afero.ReadFile(fileSystem, machineIdPath)
	if err != nil {
		return "", err
	}

	machineID := strings.TrimSpace(string(machineIDBytes))
	agentID := uuid.NewSHA1(trentoNamespace, []byte(machineID))

	return agentID.String(), nil
}

// Start the Agent. This will start the discovery ticker and the heartbeat ticker
func (a *Agent) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	for _, d := range a.discoveries {
		wg.Add(1)
		go func(wg *sync.WaitGroup, d discovery.Discovery) {
			log.Infof("Starting %s loop...", d.GetId())
			defer wg.Done()
			a.startDiscoverTicker(ctx, d)
			log.Infof("%s discover loop stopped.", d.GetId())
		}(&wg, d)
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Info("Starting heartbeat loop...")
		defer wg.Done()
		a.startHeartbeatTicker(ctx)
		log.Info("heartbeat loop stopped.")
	}(&wg)

	if a.config.FactsEngineEnabled {
		wg.Add(1)
		c := factsengine.NewFactsEngine(a.agentID, a.config.FactsServiceUrl)
		go func(wg *sync.WaitGroup) {
			log.Info("Starting fact gathering service...")
			defer wg.Done()
			c.Subscribe()
			c.Listen(ctx)
			log.Info("fact gathering stopped.")
		}(&wg)
	}

	wg.Wait()

	return nil
}

func (a *Agent) Stop(ctxCancel context.CancelFunc) {
	ctxCancel()
}

// Start a Ticker loop that will iterate over the hardcoded list of Discovery backends and execute them.
func (a *Agent) startDiscoverTicker(ctx context.Context, d discovery.Discovery) {

	tick := func() {
		result, err := d.Discover()
		if err != nil {
			result = fmt.Sprintf("Error while running discovery '%s': %s", d.GetId(), err)
			log.Errorln(result)
		}
		log.Infof("%s discovery tick output: %s", d.GetId(), result)
	}
	Repeat(d.GetId(), tick, time.Duration(d.GetInterval()), ctx)

}

func (a *Agent) startHeartbeatTicker(ctx context.Context) {
	tick := func() {
		err := a.collectorClient.Heartbeat()
		if err != nil {
			log.Errorf("Error while sending the heartbeat to the server: %s", err)
		}
	}

	Repeat("agent.heartbeat", tick, HeartbeatInterval, ctx)
}
