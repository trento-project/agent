package cmd

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/trento-project/agent/internal/agent"
	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/collector"
)

const prometheusModePush = "push"

func buildPushModePrometheusConfig() (*discovery.PrometheusConfig, error) {
	prometheusURL := viper.GetString("prometheus-url")
	if prometheusURL == "" {
		return nil, errors.New("prometheus-url is required when prometheus-mode is 'push'")
	}

	exporterName := viper.GetString("prometheus-exporter-name")
	if exporterName == "" {
		exporterName = "grafana_alloy"
	}

	return &discovery.PrometheusConfig{
		Mode:         prometheusModePush,
		Target:       prometheusURL,
		ExporterName: exporterName,
	}, nil
}

func buildPullModePrometheusConfig() *discovery.PrometheusConfig {
	target := viper.GetString("prometheus-node-exporter-target")
	if target == "" {
		target = viper.GetString("node-exporter-target")
	}

	exporterName := viper.GetString("prometheus-exporter-name")
	if exporterName == "" {
		exporterName = viper.GetString("node-exporter-name")
		if exporterName == "" {
			exporterName = "node_exporter"
		}
	}

	return &discovery.PrometheusConfig{
		Mode:         viper.GetString("prometheus-mode"),
		Target:       target,
		ExporterName: exporterName,
	}
}

func validatePeriod(durationFlag string, minValue time.Duration) error {
	period := viper.GetDuration(durationFlag)
	if period < minValue {
		return errors.Errorf("%s: invalid interval %s, should be at least %s", durationFlag, period, minValue)
	}

	return nil
}

func LoadConfig(fileSystem afero.Fs) (*agent.Config, error) {
	minPeriodValues := map[string]time.Duration{
		"cluster-discovery-period":      discovery.ClusterDiscoveryMinPeriod,
		"sapsystem-discovery-period":    discovery.SAPDiscoveryMinPeriod,
		"cloud-discovery-period":        discovery.CloudDiscoveryMinPeriod,
		"host-discovery-period":         discovery.HostDiscoveryMinPeriod,
		"subscription-discovery-period": discovery.SubscriptionDiscoveryMinPeriod,
		"saptune-discovery-period":      discovery.SaptuneDiscoveryMinPeriod,
	}

	for flagName, minPeriodValue := range minPeriodValues {
		err := validatePeriod(flagName, minPeriodValue)
		if err != nil {
			return nil, err
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "could not read the hostname")
	}

	apiKey := viper.GetString("api-key")
	if apiKey == "" {
		return nil, errors.New("api-key is required, cannot start agent")
	}

	agentID := viper.GetString("force-agent-id")
	if agentID == "" {
		id, err := agent.GetAgentID(fileSystem)
		if err != nil {
			return nil, errors.Wrap(err, "could not get the agent ID")
		}
		agentID = id
	}

	collectorConfig := &collector.Config{
		ServerURL: viper.GetString("server-url"),
		APIKey:    apiKey,
		AgentID:   agentID,
	}

	discoveryPeriodsConfig := &discovery.DiscoveriesPeriodConfig{
		Cluster:      viper.GetDuration("cluster-discovery-period"),
		SAPSystem:    viper.GetDuration("sapsystem-discovery-period"),
		Cloud:        viper.GetDuration("cloud-discovery-period"),
		Host:         viper.GetDuration("host-discovery-period"),
		Subscription: viper.GetDuration("subscription-discovery-period"),
		Saptune:      viper.GetDuration("saptune-discovery-period"),
	}

	discoveriesConfig := &discovery.DiscoveriesConfig{
		CollectorConfig:          collectorConfig,
		DiscoveriesPeriodsConfig: discoveryPeriodsConfig,
	}

	var prometheusConfig *discovery.PrometheusConfig
	if viper.GetString("prometheus-mode") == prometheusModePush {
		var err error
		prometheusConfig, err = buildPushModePrometheusConfig()
		if err != nil {
			return nil, err
		}
	} else {
		prometheusConfig = buildPullModePrometheusConfig()
	}

	return &agent.Config{
		AgentID:           agentID,
		InstanceName:      hostname,
		DiscoveriesConfig: discoveriesConfig,
		FactsServiceURL:   viper.GetString("facts-service-url"),
		PluginsFolder:     viper.GetString("plugins-folder"),
		PrometheusConfig:  prometheusConfig,
	}, nil
}
