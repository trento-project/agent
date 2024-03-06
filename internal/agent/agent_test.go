package agent_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/agent"
	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/test/helpers"
)

type AgentTestSuite struct {
	suite.Suite
}

func TestAgentTestSuite(t *testing.T) {
	suite.Run(t, new(AgentTestSuite))
}

func (suite *AgentTestSuite) TestAgentGetAgentID() {
	fileSystem := helpers.MockMachineIDFile()
	agentID, err := agent.GetAgentID(fileSystem)

	suite.NoError(err)
	suite.Equal(helpers.DummyAgentID, agentID)
}

func (suite *AgentTestSuite) TestAgentFailsWithInvalidFactsServiceURL() {
	config := &agent.Config{
		AgentID:      helpers.DummyAgentID,
		InstanceName: "test",
		DiscoveriesConfig: &discovery.DiscoveriesConfig{
			DiscoveriesPeriodsConfig: &discovery.DiscoveriesPeriodConfig{},
			CollectorConfig:          &collector.Config{},
		},
		FactsServiceURL: "amqp://trento:trento@somehost:1234/somevhost",
		PluginsFolder:   "/usr/etc/trento/plugins/",
	}

	agent, _ := agent.NewAgent(config)
	ctx := context.Background()
	err := agent.Start(ctx)

	suite.ErrorContains(err, "dial tcp: lookup somehost: no such host")
}
