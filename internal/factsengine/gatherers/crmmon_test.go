package gatherers // nolint

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
	mocks "github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type CrmMonTestSuite struct {
	suite.Suite
	mockExecutor *mocks.CommandExecutor
	crmMonOutput []byte
}

func TestCrmMonTestSuite(t *testing.T) {
	suite.Run(t, new(CrmMonTestSuite))
}

func (suite *CrmMonTestSuite) SetupTest() {
	suite.mockExecutor = new(mocks.CommandExecutor)
	lFile, _ := os.Open("../../../test/fixtures/gatherers/crmmon.xml")
	content, _ := io.ReadAll(lFile)

	suite.crmMonOutput = content
}

func (suite *CrmMonTestSuite) TestCrmMonGather() {
	suite.mockExecutor.On("Exec", "crm_mon", "--output-as", "xml").Return(
		suite.crmMonOutput, nil)

	p := NewCrmMonGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "role",
			Gatherer: "crm_mon",
			Argument: "//resource[@resource_agent='stonith:external/sbd']/@role",
			CheckID:  "check1",
		},
		{
			Name:     "active",
			Gatherer: "crm_mon",
			Argument: "//resource[@resource_agent='stonith:external/sbd']/@active",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "role",
			Value:   "Started",
			CheckID: "check1",
		},
		{
			Name:    "active",
			Value:   "true",
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CrmMonTestSuite) TestCrmMonGatherCmdNotFound() {
	suite.mockExecutor.On("Exec", "crm_mon", "--output-as", "xml").Return(
		suite.crmMonOutput, errors.New("crm_mon not found"))
	p := NewCrmMonGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "role",
			Gatherer: "crm_mon",
			Argument: "//resource[@resource_agent='stonith:external/sbd']/@rle",
			CheckID:  "check1",
		},
		{
			Name:     "active",
			Gatherer: "crm_mon",
			Argument: "//resource[@resource_agent='stonith:external/sbd']/@active",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "role",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "error running crm_mon command: crm_mon not found",
				Type:    "crmmon-execution-error",
			},
		},
		{
			Name:    "active",
			Value:   nil,
			CheckID: "check2",
			Error: &entities.FactGatheringError{
				Message: "error running crm_mon command: crm_mon not found",
				Type:    "crmmon-execution-error",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CrmMonTestSuite) TestCrmMonGatherError() {

	suite.mockExecutor.On("Exec", "crm_mon", "--output-as", "xml").Return(
		suite.crmMonOutput, nil)

	p := NewCrmMonGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "role",
			Gatherer: "crm_mon",
			Argument: "//resource[@resource_agent='stonith:external/sbd']/@rle",
			CheckID:  "check1",
		},
		{
			Name:     "active",
			Gatherer: "crm_mon",
			Argument: "//resource[@resource_agent='stonith:external/sbd']/@active",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "role",
			Value:   nil,
			CheckID: "check1",
			Error: &entities.FactGatheringError{
				Message: "requested xpath value not found: " +
					"requested xpath //resource[@resource_agent='stonith:external/sbd']/@rle not found",
				Type: "xml-xpath-value-not-found",
			},
		},
		{
			Name:    "active",
			Value:   "true",
			CheckID: "check2",
			Error:   nil,
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
