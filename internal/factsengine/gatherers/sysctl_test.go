// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package gatherers_test

import (
	"context"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/v3/internal/factsengine/gatherers"
	"github.com/trento-project/agent/v3/pkg/factsengine/entities"
	"github.com/trento-project/agent/v3/pkg/utils"
	utilsMocks "github.com/trento-project/agent/v3/pkg/utils/mocks"
	"github.com/trento-project/agent/v3/test/helpers"
)

type SysctlTestSuite struct {
	suite.Suite

	mockExecutor *utilsMocks.MockCommandExecutor
}

func TestSysctlTestSuite(t *testing.T) {
	suite.Run(t, new(SysctlTestSuite))
}

func (suite *SysctlTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.MockCommandExecutor)
}

func (suite *SysctlTestSuite) TestSysctlGathererNoArgumentProvided() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/sysctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("OutputContext", mock.Anything, "/sbin/sysctl", "-a").Return(mockOutput, nil)

	c := gatherers.NewSysctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "sysctl",
		},
		{
			Name:     "empty_argument_fact",
			Gatherer: "sysctl",
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
				Type:    "sysctl-missing-argument",
			},
		},
		{
			Name:  "empty_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "sysctl-missing-argument",
			},
		},
	}

	suite.Require().NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SysctlTestSuite) TestSysctlGathererNonExistingKey() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/sysctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("OutputContext", mock.Anything, "/sbin/sysctl", "-a").Return(mockOutput, nil)

	c := gatherers.NewSysctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "madeup_fact",
			Gatherer: "sysctl",
			Argument: "madeup.fact",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "madeup_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "requested value not found in sysctl output: madeup.fact",
				Type:    "sysctl-value-not-found",
			},
		},
	}

	suite.Require().NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SysctlTestSuite) TestSysctlCommandNotFound() {
	suite.mockExecutor.On("OutputContext", mock.Anything, "/sbin/sysctl", "-a").Return(nil, exec.ErrNotFound)

	c := gatherers.NewSysctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "fs.inotify.max_user_watches",
			Gatherer: "sysctl",
			Argument: "fs.inotify.max_user_watches",
		},
	}

	expectedError := &entities.FactGatheringError{
		Message: "error executing sysctl command: executable file not found in $PATH",
		Type:    "sysctl-cmd-error",
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	suite.Require().EqualError(err, expectedError.Error())

	suite.Empty(factResults)
}

func (suite *SysctlTestSuite) TestSysctlGatherer() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/sysctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("OutputContext", mock.Anything, "/sbin/sysctl", "-a").Return(mockOutput, nil)

	c := gatherers.NewSysctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "simple_value",
			Gatherer: "sysctl",
			Argument: "fs.inotify.max_user_watches",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "simple_value",
			Value: &entities.FactValueInt{Value: 65536},
		},
	}

	suite.Require().NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SysctlTestSuite) TestSysctlGathererPartialKey() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/sysctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("OutputContext", mock.Anything, "/sbin/sysctl", "-a").Return(mockOutput, nil)

	c := gatherers.NewSysctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "partial_key",
			Gatherer: "sysctl",
			Argument: "debug",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "partial_key",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"exception-trace":      &entities.FactValueInt{Value: 1},
					"kprobes-optimization": &entities.FactValueInt{Value: 1},
				},
			},
		},
	}

	suite.Require().NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SysctlTestSuite) TestSysctlGathererEmptyValue() {
	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/sysctl.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("OutputContext", mock.Anything, "/sbin/sysctl", "-a").Return(mockOutput, nil)

	c := gatherers.NewSysctlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "empty_value",
			Gatherer: "sysctl",
			Argument: "kernel.domainname",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "empty_value",
			Value: &entities.FactValueString{Value: ""},
		},
	}

	suite.Require().NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SysctlTestSuite) TestSysctlGathererContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := gatherers.NewSysctlGatherer(utils.Executor{})
	factRequests := []entities.FactRequest{
		{
			Name:     "context_cancelled_fact",
			Gatherer: "sysctl",
		},
	}
	factResults, err := c.Gather(ctx, factRequests)

	suite.Require().Error(err)
	suite.Empty(factResults)
}
