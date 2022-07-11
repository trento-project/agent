package internal

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	_ "github.com/trento-project/agent/test"
)

const (
	DummyMachineID = "dummy-machine-id"
	DummyAgentID   = "779cdd70-e9e2-58ca-b18a-bf3eb3f71244"
)

type AgentTestSuite struct {
	suite.Suite
}

func TestAgentTestSuite(t *testing.T) {
	suite.Run(t, new(AgentTestSuite))
}

func (suite *AgentTestSuite) SetupSuite() {
	fileSystem = afero.NewMemMapFs()

	afero.WriteFile(fileSystem, machineIdPath, []byte(DummyMachineID), 0644)
}

func (suite *AgentTestSuite) TestAgent_getAgentID() {
	agentID, err := getAgentID()

	suite.NoError(err)
	suite.Equal(DummyAgentID, agentID)
}
