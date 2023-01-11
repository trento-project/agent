package gatherers_test

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type CorosyncCmapctlTestSuite struct {
	suite.Suite
	mockExecutor *utilsMocks.CommandExecutor
}

func TestCorosyncCmapctlTestSuite(t *testing.T) {
	suite.Run(t, new(CorosyncCmapctlTestSuite))
}

func (suite *CorosyncCmapctlTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
}

func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlGathererNoArgumentProvided() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/corosynccmap-ctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "corosync-cmapctl", "-b").Return(mockOutput, nil)

	c := gatherers.NewCorosyncCmapctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "corosync-cmapctl",
		},
		{
			Name:     "empty_argument_fact",
			Gatherer: "corosync-cmapctl",
			Argument: "",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "no_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "corosync-cmapctl-missing-argument",
			},
		},
		{
			Name:  "empty_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "corosync-cmapctl-missing-argument",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlGathererNonExistingKey() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/corosynccmap-ctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "corosync-cmapctl", "-b").Return(mockOutput, nil)

	c := gatherers.NewCorosyncCmapctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "madeup_fact",
			Gatherer: "corosync-cmapctl",
			Argument: "madeup.fact",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "madeup_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "requested value not found in corosync-cmapctl output: madeup.fact",
				Type:    "corosync-cmapctl-value-not-found",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlCommandNotFound() {
	suite.mockExecutor.On("Exec", "corosync-cmapctl", "-b").Return(nil, exec.ErrNotFound)

	c := gatherers.NewCorosyncCmapctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "quorum_provider",
			Gatherer: "corosync-cmapctl",
			Argument: "quorum.provider",
		},
	}

	expectedError := &entities.FactGatheringError{
		Message: "error while executing corosynccmap-ctl: executable file not found in $PATH",
		Type:    "corosync-cmapctl-command-error",
	}

	factResults, err := c.Gather(factRequests)

	suite.EqualError(err, expectedError.Error())

	suite.Empty(factResults)
}

func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlGatherer() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/corosynccmap-ctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "corosync-cmapctl", "-b").Return(mockOutput, nil)

	c := gatherers.NewCorosyncCmapctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "nodes",
			Gatherer: "corosync-cmapctl",
			Argument: "nodelist.node",
		},
		{
			Name:     "totem_max_messages",
			Gatherer: "corosync-cmapctl",
			Argument: "runtime.config.totem.max_messages",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "nodes",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"0": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"nodeid":     &entities.FactValueInt{Value: 1},
							"ring0_addr": &entities.FactValueString{Value: "10.80.1.11"},
						},
					},
					"1": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"nodeid":     &entities.FactValueInt{Value: 2},
							"ring0_addr": &entities.FactValueString{Value: "10.80.1.12"},
						},
					},
				},
			},
		},
		{
			Name:  "totem_max_messages",
			Value: &entities.FactValueInt{Value: 20},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
