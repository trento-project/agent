package cmd

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal"
	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/collector"
)

type AgentCmdTestSuite struct {
	suite.Suite
	cmd *cobra.Command
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
}

func (suite *AgentCmdTestSuite) TearDownTest() {
	_ = suite.cmd.Execute()

	expectedConfig := &internal.Config{
		InstanceName: "some-hostname",
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
				AgentID:   "",
			},
		},
		FactsEngineEnabled: false,
		FactsServiceURL:    "amqp://guest:guest@localhost:5672",
		PluginsFolder:      "/usr/etc/trento/plugins/",
	}

	config, err := LoadConfig()
	config.InstanceName = "some-hostname"
	suite.NoError(err)

	suite.EqualValues(expectedConfig, config)
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
	})
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
}

func (suite *AgentCmdTestSuite) TestConfigFromFile() {
	os.Setenv("TRENTO_CONFIG", "../test/fixtures/config/agent.yaml")
}
