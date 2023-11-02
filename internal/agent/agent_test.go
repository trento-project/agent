package agent_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/agent"
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
