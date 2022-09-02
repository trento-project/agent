package gatherers

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
	mocks "github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type CorosyncCmapctlTestSuite struct {
	suite.Suite
	mockExecutor *mocks.CommandExecutor
}

func TestCorosyncCmapctlTestSuite(t *testing.T) {
	suite.Run(t, new(CorosyncCmapctlTestSuite))
}

func (suite *CorosyncCmapctlTestSuite) SetupTest() {
	suite.mockExecutor = new(mocks.CommandExecutor)
}

func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlGathererMissingFact() {
	mockOutputFile, _ := os.Open("../../../test/fixtures/gatherers/corosynccmap-ctl.output")
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "corosync-cmapctl", "-b").Return(mockOutput, nil)

	c := NewCorosyncCmapctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "madeup_fact",
			Gatherer: "corosync-cmapctl",
			Argument: "madeup.fact",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlGatherer() {
	mockOutputFile, _ := os.Open("../../../test/fixtures/gatherers/corosynccmap-ctl.output")
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "corosync-cmapctl", "-b").Return(mockOutput, nil)

	c := NewCorosyncCmapctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "quorum_provider",
			Gatherer: "corosync-cmapctl",
			Argument: "quorum.provider",
		},
		{
			Name:     "totem_max_messages",
			Gatherer: "corosync-cmapctl",
			Argument: "runtime.config.totem.max_messages",
		},
		{
			Name:     "totem_transport",
			Gatherer: "corosync-cmapctl",
			Argument: "totem.transport",
		},
		{
			Name:     "votequorum_two_node",
			Gatherer: "corosync-cmapctl",
			Argument: "runtime.votequorum.two_node",
		},
		{
			Name:     "totem_consensus",
			Gatherer: "corosync-cmapctl",
			Argument: "runtime.config.totem.consensus",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:  "quorum_provider",
			Value: "corosync_votequorum",
		},
		{
			Name:  "totem_max_messages",
			Value: "20",
		},
		{
			Name:  "totem_transport",
			Value: "udpu",
		},
		{
			Name:  "votequorum_two_node",
			Value: "1",
		},
		{
			Name:  "totem_consensus",
			Value: "36000",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlCommandNotFound() {
	suite.mockExecutor.On("Exec", "corosync-cmapctl", "-b").Return(nil, exec.ErrNotFound)

	c := NewCorosyncCmapctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "quorum_provider",
			Gatherer: "corosync-cmapctl",
			Argument: "quorum.provider",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{}

	suite.Error(err)
	suite.ElementsMatch(expectedResults, factResults)
}
