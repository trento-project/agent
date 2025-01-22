package gatherers_test

import (
	"context"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/mock"
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

// nolint:dupl
func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlGathererNoArgumentProvided() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/corosynccmap-ctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("ExecContext", mock.Anything, "corosync-cmapctl", "-b").Return(mockOutput, nil)

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

	factResults, err := c.Gather(context.Background(), factRequests)

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
	suite.mockExecutor.On("ExecContext", mock.Anything, "corosync-cmapctl", "-b").Return(mockOutput, nil)

	c := gatherers.NewCorosyncCmapctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "madeup_fact",
			Gatherer: "corosync-cmapctl",
			Argument: "madeup.fact",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

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
	suite.mockExecutor.On("ExecContext", mock.Anything, "corosync-cmapctl", "-b").Return(nil, exec.ErrNotFound)

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

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.EqualError(err, expectedError.Error())

	suite.Empty(factResults)
}

func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlGatherer() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/corosynccmap-ctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("ExecContext", mock.Anything, "corosync-cmapctl", "-b").Return(mockOutput, nil)

	c := gatherers.NewCorosyncCmapctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "simple_value",
			Gatherer: "corosync-cmapctl",
			Argument: "runtime.config.totem.max_messages",
		},
		{
			Name:     "brackets_in_value_field",
			Gatherer: "corosync-cmapctl",
			Argument: "runtime.totem.pg.mrp.srp.members.1.ip",
		},
		{
			Name:     "map_of_nodes",
			Gatherer: "corosync-cmapctl",
			Argument: "nodelist.node",
		},
		{
			Name:     "nested_map_and_primitive_value",
			Gatherer: "corosync-cmapctl",
			Argument: "runtime.services.cmap",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "simple_value",
			Value: &entities.FactValueInt{Value: 20},
		},
		{
			Name:  "brackets_in_value_field",
			Value: &entities.FactValueString{Value: "r(0) ip(10.80.1.11) "},
		},
		{
			Name: "map_of_nodes",
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
			Name: "nested_map_and_primitive_value",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"0": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"rx": &entities.FactValueInt{Value: 3},
							"tx": &entities.FactValueInt{Value: 2},
						},
					},
					"service_id": &entities.FactValueInt{Value: 0},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CorosyncCmapctlTestSuite) TestCorosyncCmapctlGathererContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	suite.mockExecutor.
		On("ExecContext", mock.Anything, "corosync-cmapctl", "-b").
		Return(nil, ctx.Err())

	c := gatherers.NewCorosyncCmapctlGatherer(suite.mockExecutor)
	factRequests := []entities.FactRequest{
		{
			Name:     "madeup_fact",
			Gatherer: "corosync-cmapctl",
			Argument: "madeup.fact",
		},
	}
	factResults, err := c.Gather(ctx, factRequests)

	suite.Error(err)
	suite.Empty(factResults)
}
