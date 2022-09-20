package factsengine

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

const (
	executionID = "someExecution"
	agentID     = "someAgent"
)

type GatheringTestSuite struct {
	suite.Suite
}

func TestGatheringTestSuite(t *testing.T) {
	suite.Run(t, new(GatheringTestSuite))
}

type DummyGatherer1 struct {
}

func NewDummyGatherer1() *DummyGatherer1 {
	return &DummyGatherer1{}
}

func (s *DummyGatherer1) Gather(_ []entities.FactRequest) ([]entities.Fact, error) {
	return []entities.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
	}, nil
}

type DummyGatherer2 struct {
}

func NewDummyGatherer2() *DummyGatherer2 {
	return &DummyGatherer2{}
}

func (s *DummyGatherer2) Gather(_ []entities.FactRequest) ([]entities.Fact, error) {
	return []entities.Fact{
		{
			Name:    "dummy2",
			Value:   "2",
			CheckID: "check1",
		},
	}, nil
}

type ErrorGatherer struct {
}

func NewErrorGatherer() *ErrorGatherer {
	return &ErrorGatherer{}
}

func (s *ErrorGatherer) Gather(_ []entities.FactRequest) ([]entities.Fact, error) {
	return nil, &entities.FactGatheringError{Type: "dummy-type", Message: "some error"}
}

func (suite *GatheringTestSuite) TestGatheringGatherFacts() {
	factsRequest := entities.FactsGatheringRequestedTarget{
		AgentID: agentID,
		FactRequests: []entities.FactRequest{
			{
				Name:     "dummy1",
				Gatherer: "dummyGatherer1",
				Argument: "dummy1",
				CheckID:  "check1",
			},
			{
				Name:     "dummy2",
				Gatherer: "dummyGatherer2",
				Argument: "dummy2",
				CheckID:  "check1",
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"dummyGatherer2": NewDummyGatherer2(),
	}

	factResults, err := gatherFacts(executionID, agentID, &factsRequest, factGatherers)

	expectedFacts := []entities.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
		{
			Name:    "dummy2",
			Value:   "2",
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(executionID, factResults.ExecutionID)
	suite.Equal(agentID, factResults.AgentID)
	suite.ElementsMatch(expectedFacts, factResults.FactsGathered)
}

func (suite *GatheringTestSuite) TestFactsEngineGatherFactsGathererNotFound() {
	factsRequest := entities.FactsGatheringRequestedTarget{
		AgentID: agentID,
		FactRequests: []entities.FactRequest{
			{
				Name:     "dummy1",
				Gatherer: "dummyGatherer1",
				Argument: "dummy1",
				CheckID:  "check1",
			},
			{
				Name:     "other",
				Gatherer: "otherGatherer",
				Argument: "other",
				CheckID:  "check1",
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"dummyGatherer2": NewDummyGatherer2(),
	}

	factResults, err := gatherFacts(executionID, agentID, &factsRequest, factGatherers)

	expectedFacts := []entities.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(executionID, factResults.ExecutionID)
	suite.Equal(agentID, factResults.AgentID)
	suite.ElementsMatch(expectedFacts, factResults.FactsGathered)
}

func (suite *GatheringTestSuite) TestFactsEngineGatherFactsErrorGathering() {
	factsRequest := entities.FactsGatheringRequestedTarget{
		AgentID: agentID,
		FactRequests: []entities.FactRequest{
			{
				Name:     "dummy1",
				Gatherer: "dummyGatherer1",
				Argument: "dummy1",
				CheckID:  "check1",
			},
			{
				Name:     "error",
				Gatherer: "errorGatherer",
				Argument: "error",
				CheckID:  "check1",
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"errorGatherer":  NewErrorGatherer(),
	}

	factResults, err := gatherFacts(executionID, agentID, &factsRequest, factGatherers)

	expectedFacts := []entities.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
			Error:   nil,
		},
		{
			Name:    "error",
			Value:   nil,
			CheckID: "check1",
			Error: &entities.FactGatheringError{
				Type:    "dummy-type",
				Message: "some error",
			},
		},
	}

	suite.NoError(err)
	suite.Equal(executionID, factResults.ExecutionID)
	suite.Equal(agentID, factResults.AgentID)
	suite.ElementsMatch(expectedFacts, factResults.FactsGathered)
}
