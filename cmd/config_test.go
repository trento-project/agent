package cmd

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/agent"
	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/test/helpers"
)

const hostname = "some-hostname"

// nolint
var expectedConfig = &agent.Config{
	AgentID:      "some-agent-id",
	InstanceName: hostname,
	DiscoveriesConfig: &discovery.DiscoveriesConfig{
		SSHAddress: "some-ssh-address",
		DiscoveriesPeriodsConfig: &discovery.DiscoveriesPeriodConfig{
			Cluster:      10 * time.Second,
			SAPSystem:    10 * time.Second,
			Cloud:        10 * time.Second,
			Host:         10 * time.Second,
			Subscription: 900 * time.Second,
		},
		CollectorConfig: &collector.Config{
			ServerURL: "http://serverurl",
			APIKey:    "some-api-key",
			AgentID:   "some-agent-id",
		},
	},
	FactsEngineEnabled: false,
	FactsServiceURL:    "amqp://guest:guest@localhost:5672",
	PluginsFolder:      "/usr/etc/trento/plugins/",
}

type AgentCmdTestSuite struct {
	suite.Suite
	cmd        *cobra.Command
	fileSystem afero.Fs
}

func TestAgentCmdTestSuite(t *testing.T) {
	suite.Run(t, new(AgentCmdTestSuite))
}

func (suite *AgentCmdTestSuite) SetupTest() {
	os.Clearenv()

	cmd := NewRootCmd()

	for _, command := range cmd.Commands() {
		command.Run = func(cmd *cobra.Command, args []string) {
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
}

func (suite *AgentCmdTestSuite) TestConfigFromFlags() {
	suite.cmd.SetArgs([]string{
		"start",
		"--ssh-address=some-ssh-address",
		"--cloud-discovery-period=10s",
		"--cluster-discovery-period=10s",
		"--sapsystem-discovery-period=10s",
		"--host-discovery-period=10s",
		"--subscription-discovery-period=900s",
		"--server-url=http://serverurl",
		"--api-key=some-api-key",
		"--agent-id=some-agent-id",
	})

	_ = suite.cmd.Execute()

	config, err := LoadConfig(suite.fileSystem)
	config.InstanceName = hostname
	suite.NoError(err)

	suite.EqualValues(expectedConfig, config)
}

func (suite *AgentCmdTestSuite) TestConfigFromEnv() {
	os.Setenv("TRENTO_SSH_ADDRESS", "some-ssh-address")
	os.Setenv("TRENTO_CLOUD_DISCOVERY_PERIOD", "10s")
	os.Setenv("TRENTO_CLUSTER_DISCOVERY_PERIOD", "10s")
	os.Setenv("TRENTO_SAPSYSTEM_DISCOVERY_PERIOD", "10s")
	os.Setenv("TRENTO_HOST_DISCOVERY_PERIOD", "10s")
	os.Setenv("TRENTO_SUBSCRIPTION_DISCOVERY_PERIOD", "900s")
	os.Setenv("TRENTO_SERVER_URL", "http://serverurl")
	os.Setenv("TRENTO_API_KEY", "some-api-key")
	os.Setenv("TRENTO_AGENT_ID", "some-agent-id")

	_ = suite.cmd.Execute()

	config, err := LoadConfig(suite.fileSystem)
	config.InstanceName = hostname
	suite.NoError(err)

	suite.EqualValues(expectedConfig, config)
}

func (suite *AgentCmdTestSuite) TestConfigFromFile() {
	os.Setenv("TRENTO_CONFIG", "../test/fixtures/config/agent.yaml")

	_ = suite.cmd.Execute()

	config, err := LoadConfig(suite.fileSystem)
	config.InstanceName = hostname
	suite.NoError(err)

	suite.EqualValues(expectedConfig, config)
}

func (suite *AgentCmdTestSuite) TestAgentIDLoaded() {
	suite.cmd.SetArgs([]string{
		"start",
		"--ssh-address=some-ssh-address",
		"--api-key=some-api-key",
	})

	_ = suite.cmd.Execute()

	config, err := LoadConfig(suite.fileSystem)
	suite.NoError(err)
	suite.Equal(helpers.DummyAgentID, config.AgentID)
	suite.Equal(helpers.DummyAgentID, config.DiscoveriesConfig.CollectorConfig.AgentID)
}
