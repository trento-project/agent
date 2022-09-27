package factsengine

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
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

	dummyGathererOne := &mocks.FactGatherer{}
	dummyGathererOne.On("Gather", mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy1",
				Value:   "1",
				CheckID: "check1",
			},
		}, nil).Times(1)

	dummyGathererTwo := &mocks.FactGatherer{}
	dummyGathererTwo.On("Gather", mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy2",
				Value:   "2",
				CheckID: "check1",
			},
		}, nil).Times(1)

	registry := gatherers.NewRegistry(map[string]gatherers.FactGatherer{
		"dummyGatherer1": dummyGathererOne,
		"dummyGatherer2": dummyGathererTwo,
	})

	factResults, err := gatherFacts(executionID, agentID, &factsRequest, *registry)

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

	dummyGathererOne := &mocks.FactGatherer{}
	dummyGathererOne.On("Gather", mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy1",
				Value:   "1",
				CheckID: "check1",
			},
		}, nil).Times(1)

	dummyGathererTwo := &mocks.FactGatherer{}
	dummyGathererTwo.On("Gather", mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy2",
				Value:   "2",
				CheckID: "check1",
			},
		}, nil).Times(1)

	registry := gatherers.NewRegistry(map[string]gatherers.FactGatherer{
		"dummyGatherer1": dummyGathererOne,
		"dummyGatherer2": dummyGathererTwo,
	})

	factResults, err := gatherFacts(executionID, agentID, &factsRequest, *registry)

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

	dummyGathererOne := &mocks.FactGatherer{}
	dummyGathererOne.On("Gather", mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy1",
				Value:   "1",
				CheckID: "check1",
			},
		}, nil).Times(1)

	errorGatherer := &mocks.FactGatherer{}
	errorGatherer.On("Gather", mock.Anything).
		Return(nil, &entities.FactGatheringError{Type: "dummy-type", Message: "some error"}).Times(1)

	registry := gatherers.NewRegistry(map[string]gatherers.FactGatherer{
		"dummyGatherer1": dummyGathererOne,
		"errorGatherer":  errorGatherer,
	})

	factResults, err := gatherFacts(executionID, agentID, &factsRequest, *registry)

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
