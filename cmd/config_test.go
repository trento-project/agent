package cmd_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/cmd"
	"github.com/trento-project/agent/internal/agent"
	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/test/helpers"
)

type AgentCmdTestSuite struct {
	suite.Suite
	cmd            *cobra.Command
	fileSystem     afero.Fs
	hostname       string
	expectedConfig *agent.Config
}

func TestAgentCmdTestSuite(t *testing.T) {
	suite.Run(t, new(AgentCmdTestSuite))
}

func (suite *AgentCmdTestSuite) SetupTest() {
	os.Clearenv()
	viper.Reset()

	cmd := cmd.NewRootCmd()

	for _, command := range cmd.Commands() {
		command.Run = func(_ *cobra.Command, _ []string) {
			// do nothing
		}
	}

	cmd.SetArgs([]string{
		"start",
	})

	var b bytes.Buffer
	cmd.SetOut(&b)

	suite.cmd = cmd
	suite.fileSystem = helpers.MockMachineIDFile()
	suite.hostname = "some-hostname"
	suite.expectedConfig = &agent.Config{
		AgentID:      "some-agent-id",
		InstanceName: "some-hostname",
		DiscoveriesConfig: &discovery.DiscoveriesConfig{
			DiscoveriesPeriodsConfig: &discovery.DiscoveriesPeriodConfig{
				Cluster:      10 * time.Second,
				SAPSystem:    10 * time.Second,
				Cloud:        10 * time.Second,
				Host:         10 * time.Second,
				Subscription: 900 * time.Second,
				Saptune:      900 * time.Second,
			},
			CollectorConfig: &collector.Config{
				ServerURL: "http://serverurl",
				APIKey:    "some-api-key",
				AgentID:   "some-agent-id",
			},
		},
		FactsServiceURL: "amqp://guest:guest@serviceurl:5672",
		PluginsFolder:   "/usr/etc/trento/plugins/",
		PrometheusConfig: &discovery.PrometheusConfig{
			Mode:         "pull",
			ExporterName: "node_exporter",
			Target:       "10.0.0.5:9100",
		},
	}
}

func (suite *AgentCmdTestSuite) TestConfigFromFlags() {
	suite.cmd.SetArgs([]string{
		"start",
		"--cloud-discovery-period=10s",
		"--cluster-discovery-period=10s",
		"--sapsystem-discovery-period=10s",
		"--host-discovery-period=10s",
		"--subscription-discovery-period=900s",
		"--saptune-discovery-period=900s",
		"--server-url=http://serverurl",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--facts-service-url=amqp://guest:guest@serviceurl:5672",
		"--node-exporter-target=10.0.0.5:9100",
	})

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	config.InstanceName = suite.hostname
	suite.NoError(err)

	suite.EqualValues(suite.expectedConfig, config)
}

func (suite *AgentCmdTestSuite) TestConfigFromEnv() {
	os.Setenv("TRENTO_CLOUD_DISCOVERY_PERIOD", "10s")
	os.Setenv("TRENTO_CLUSTER_DISCOVERY_PERIOD", "10s")
	os.Setenv("TRENTO_SAPSYSTEM_DISCOVERY_PERIOD", "10s")
	os.Setenv("TRENTO_HOST_DISCOVERY_PERIOD", "10s")
	os.Setenv("TRENTO_SUBSCRIPTION_DISCOVERY_PERIOD", "900s")
	os.Setenv("TRENTO_SAPTUNE_DISCOVERY_PERIOD", "900s")
	os.Setenv("TRENTO_SERVER_URL", "http://serverurl")
	os.Setenv("TRENTO_API_KEY", "some-api-key")
	os.Setenv("TRENTO_FORCE_AGENT_ID", "some-agent-id")
	os.Setenv("TRENTO_FACTS_SERVICE_URL", "amqp://guest:guest@serviceurl:5672")
	os.Setenv("TRENTO_NODE_EXPORTER_TARGET", "10.0.0.5:9100")

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	config.InstanceName = suite.hostname
	suite.NoError(err)

	suite.EqualValues(suite.expectedConfig, config)
}

func (suite *AgentCmdTestSuite) TestConfigFromFile() {
	os.Setenv("TRENTO_CONFIG", "../test/fixtures/config/agent.yaml")

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	config.InstanceName = suite.hostname
	suite.NoError(err)

	suite.EqualValues(suite.expectedConfig, config)
}

func (suite *AgentCmdTestSuite) TestAgentIDLoaded() {
	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
	})

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal(helpers.DummyAgentID, config.AgentID)
	suite.Equal(helpers.DummyAgentID, config.DiscoveriesConfig.CollectorConfig.AgentID)
}

func (suite *AgentCmdTestSuite) TestConfigPrometheusPushModeWithURL() {
	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--prometheus-mode=push",
		"--prometheus-url=http://pushgateway:9091",
	})

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal("push", config.PrometheusConfig.Mode)
	suite.Equal("http://pushgateway:9091", config.PrometheusConfig.Target)
	suite.Equal("grafana_alloy", config.PrometheusConfig.ExporterName)
}

func (suite *AgentCmdTestSuite) TestConfigPrometheusPushModeWithCustomExporterName() {
	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--prometheus-mode=push",
		"--prometheus-url=http://pushgateway:9091",
		"--prometheus-exporter-name=custom_exporter",
	})

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal("push", config.PrometheusConfig.Mode)
	suite.Equal("http://pushgateway:9091", config.PrometheusConfig.Target)
	suite.Equal("custom_exporter", config.PrometheusConfig.ExporterName)
}

func (suite *AgentCmdTestSuite) TestConfigPrometheusPushModeMissingURL() {
	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--prometheus-mode=push",
	})

	_ = suite.cmd.Execute()

	_, err := cmd.LoadConfig(suite.fileSystem)
	suite.Error(err)
	suite.Contains(err.Error(), "prometheus-url is required when prometheus-mode is 'push'")
}

func (suite *AgentCmdTestSuite) TestConfigPrometheusPullModeWithNewFlags() {
	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--prometheus-mode=pull",
		"--prometheus-node-exporter-target=10.0.0.10:9100",
		"--prometheus-exporter-name=custom_node_exporter",
	})

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal("pull", config.PrometheusConfig.Mode)
	suite.Equal("10.0.0.10:9100", config.PrometheusConfig.Target)
	suite.Equal("custom_node_exporter", config.PrometheusConfig.ExporterName)
}

func (suite *AgentCmdTestSuite) TestConfigPrometheusPullModeDefaultExporterName() {
	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--prometheus-node-exporter-target=10.0.0.10:9100",
	})

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal("node_exporter", config.PrometheusConfig.ExporterName)
}

func (suite *AgentCmdTestSuite) TestConfigLegacyNodeExporterNameFallback() {
	os.Setenv("TRENTO_NODE_EXPORTER_NAME", "legacy_exporter")

	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--node-exporter-target=10.0.0.10:9100",
	})

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal("10.0.0.10:9100", config.PrometheusConfig.Target)
	suite.Equal("legacy_exporter", config.PrometheusConfig.ExporterName)
}

func (suite *AgentCmdTestSuite) TestConfigPrometheusNodeExporterTargetOverridesLegacy() {
	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--node-exporter-target=10.0.0.5:9100",
		"--prometheus-node-exporter-target=10.0.0.10:9100",
	})

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal("10.0.0.10:9100", config.PrometheusConfig.Target)
}

func (suite *AgentCmdTestSuite) TestConfigMissingAPIKey() {
	suite.cmd.SetArgs([]string{
		"start",
		"--force-agent-id=some-agent-id",
	})

	_ = suite.cmd.Execute()

	_, err := cmd.LoadConfig(suite.fileSystem)
	suite.Error(err)
	suite.Contains(err.Error(), "api-key is required")
}

func (suite *AgentCmdTestSuite) TestConfigInvalidClusterDiscoveryPeriod() {
	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--cluster-discovery-period=500ms",
	})

	_ = suite.cmd.Execute()

	_, err := cmd.LoadConfig(suite.fileSystem)
	suite.Error(err)
	suite.Contains(err.Error(), "cluster-discovery-period")
	suite.Contains(err.Error(), "invalid interval")
}

func (suite *AgentCmdTestSuite) TestConfigInvalidSubscriptionDiscoveryPeriod() {
	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=some-api-key",
		"--force-agent-id=some-agent-id",
		"--subscription-discovery-period=10s",
	})

	_ = suite.cmd.Execute()

	_, err := cmd.LoadConfig(suite.fileSystem)
	suite.Error(err)
	suite.Contains(err.Error(), "subscription-discovery-period")
	suite.Contains(err.Error(), "invalid interval")
}

func (suite *AgentCmdTestSuite) TestConfigPrometheusPushModeFromEnv() {
	os.Setenv("TRENTO_API_KEY", "some-api-key")
	os.Setenv("TRENTO_FORCE_AGENT_ID", "some-agent-id")
	os.Setenv("TRENTO_PROMETHEUS_MODE", "push")
	os.Setenv("TRENTO_PROMETHEUS_URL", "http://pushgateway:9091")
	os.Setenv("TRENTO_PROMETHEUS_EXPORTER_NAME", "env_exporter")

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal("push", config.PrometheusConfig.Mode)
	suite.Equal("http://pushgateway:9091", config.PrometheusConfig.Target)
	suite.Equal("env_exporter", config.PrometheusConfig.ExporterName)
}

func (suite *AgentCmdTestSuite) TestConfigFlagsOverrideEnv() {
	os.Setenv("TRENTO_API_KEY", "env-api-key")
	os.Setenv("TRENTO_PROMETHEUS_MODE", "pull")

	suite.cmd.SetArgs([]string{
		"start",
		"--api-key=flag-api-key",
		"--force-agent-id=some-agent-id",
		"--prometheus-mode=push",
		"--prometheus-url=http://pushgateway:9091",
	})

	_ = suite.cmd.Execute()

	config, err := cmd.LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal("flag-api-key", config.DiscoveriesConfig.CollectorConfig.APIKey)
	suite.Equal("push", config.PrometheusConfig.Mode)
}
