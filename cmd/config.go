package cmd

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/trento-project/agent/internal"
	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/collector"
)

func validatePeriod(durationFlag string, minValue time.Duration) error {
	period := viper.GetDuration(durationFlag)
	if period < minValue {
		return errors.Errorf("%s: invalid interval %s, should be at least %s", durationFlag, period, minValue)
	}

	return nil
}

func LoadConfig() (*internal.Config, error) {
	minPeriodValues := map[string]time.Duration{
		"cluster-discovery-period":      discovery.ClusterDiscoveryMinPeriod,
		"sapsystem-discovery-period":    discovery.SAPDiscoveryMinPeriod,
		"cloud-discovery-period":        discovery.CloudDiscoveryMinPeriod,
		"host-discovery-period":         discovery.HostDiscoveryMinPeriod,
		"subscription-discovery-period": discovery.SubscriptionDiscoveryMinPeriod,
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

	sshAddress := viper.GetString("ssh-address")
	if sshAddress == "" {
		return nil, errors.New("ssh-address is required, cannot start agent")
	}

	apiKey := viper.GetString("api-key")
	if apiKey == "" {
		return nil, errors.New("api-key is required, cannot start agent")
	}

	collectorConfig := &collector.Config{
		ServerUrl: viper.GetString("server-url"),
		ApiKey:    apiKey,
	}

	discoveryPeriodsConfig := &discovery.DiscoveriesPeriodConfig{
		Cluster:      viper.GetDuration("cluster-discovery-period"),
		SAPSystem:    viper.GetDuration("sapsystem-discovery-period"),
		Cloud:        viper.GetDuration("cloud-discovery-period"),
		Host:         viper.GetDuration("host-discovery-period"),
		Subscription: viper.GetDuration("subscription-discovery-period"),
	}

	discoveriesConfig := &discovery.DiscoveriesConfig{
		SSHAddress:               sshAddress,
		CollectorConfig:          collectorConfig,
		DiscoveriesPeriodsConfig: discoveryPeriodsConfig,
	}

	return &internal.Config{
		InstanceName:      hostname,
		DiscoveriesConfig: discoveriesConfig,
		ChecksEngine:      viper.GetBool("checksengine"),
	}, nil
}
