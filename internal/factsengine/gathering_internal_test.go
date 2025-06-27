package factsengine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type GatheringTestSuite struct {
	suite.Suite
	executionID string
	agentID     string
	groupID     string
}

func TestGatheringTestSuite(t *testing.T) {
	suite.Run(t, new(GatheringTestSuite))
}

func (suite *GatheringTestSuite) SetupSuite() {
	suite.executionID = uuid.New().String()
	suite.agentID = uuid.New().String()
	suite.groupID = uuid.New().String()
}

func (suite *GatheringTestSuite) TestGatheringGatherFacts() {
	factsRequest := entities.FactsGatheringRequestedTarget{
		AgentID: suite.agentID,
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

	dummyGathererOne := &mocks.MockFactGatherer{}
	dummyGathererOne.On("Gather", mock.Anything, mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy1",
				Value:   &entities.FactValueInt{Value: 1},
				CheckID: "check1",
			},
		}, nil).Times(1)

	dummyGathererTwo := &mocks.MockFactGatherer{}
	dummyGathererTwo.On("Gather", mock.Anything, mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy2",
				Value:   &entities.FactValueInt{Value: 2},
				CheckID: "check1",
			},
		}, nil).Times(1)

	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		"dummyGatherer1": map[string]gatherers.FactGatherer{
			"v1": dummyGathererOne,
		},
		"dummyGatherer2": map[string]gatherers.FactGatherer{
			"v1": dummyGathererTwo,
		},
	})

	factResults, err := gatherFacts(context.Background(), suite.executionID, suite.agentID, suite.groupID, &factsRequest, *registry)

	expectedFacts := []entities.Fact{
		{
			Name:    "dummy1",
			Value:   &entities.FactValueInt{Value: 1},
			CheckID: "check1",
		},
		{
			Name:    "dummy2",
			Value:   &entities.FactValueInt{Value: 2},
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(suite.executionID, factResults.ExecutionID)
	suite.Equal(suite.agentID, factResults.AgentID)
	suite.Equal(suite.groupID, factResults.GroupID)
	suite.ElementsMatch(expectedFacts, factResults.FactsGathered)
}

func (suite *GatheringTestSuite) TestFactsEngineGatherFactsGathererNotFound() {
	factsRequest := entities.FactsGatheringRequestedTarget{
		AgentID: suite.agentID,
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

	dummyGathererOne := &mocks.MockFactGatherer{}
	dummyGathererOne.On("Gather", mock.Anything, mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy1",
				Value:   &entities.FactValueInt{Value: 1},
				CheckID: "check1",
			},
		}, nil).Times(1)

	dummyGathererTwo := &mocks.MockFactGatherer{}
	dummyGathererTwo.On("Gather", mock.Anything, mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy2",
				Value:   &entities.FactValueInt{Value: 1},
				CheckID: "check1",
			},
		}, nil).Times(1)

	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		"dummyGatherer1": map[string]gatherers.FactGatherer{
			"v1": dummyGathererOne,
		},
		"dummyGatherer2": map[string]gatherers.FactGatherer{
			"v1": dummyGathererTwo,
		},
	})

	factResults, err := gatherFacts(context.Background(), suite.executionID, suite.agentID, suite.groupID, &factsRequest, *registry)

	expectedFacts := []entities.Fact{
		{
			Name:    "dummy1",
			Value:   &entities.FactValueInt{Value: 1},
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(suite.executionID, factResults.ExecutionID)
	suite.Equal(suite.agentID, factResults.AgentID)
	suite.Equal(suite.groupID, factResults.GroupID)
	suite.ElementsMatch(expectedFacts, factResults.FactsGathered)
}

func (suite *GatheringTestSuite) TestFactsEngineGatherFactsErrorGathering() {
	factsRequest := entities.FactsGatheringRequestedTarget{
		AgentID: suite.agentID,
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

	dummyGathererOne := &mocks.MockFactGatherer{}
	dummyGathererOne.On("Gather", mock.Anything, mock.Anything).
		Return([]entities.Fact{
			{
				Name:    "dummy1",
				Value:   &entities.FactValueInt{Value: 1},
				CheckID: "check1",
			},
		}, nil).Times(1)

	errorGatherer := &mocks.MockFactGatherer{}
	errorGatherer.On("Gather", mock.Anything, mock.Anything).
		Return(nil, &entities.FactGatheringError{Type: "dummy-type", Message: "some error"}).Times(1)

	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		"dummyGatherer1": map[string]gatherers.FactGatherer{
			"v1": dummyGathererOne,
		},
		"errorGatherer": map[string]gatherers.FactGatherer{
			"v1": errorGatherer,
		},
	})

	factResults, err := gatherFacts(context.Background(), suite.executionID, suite.agentID, suite.groupID, &factsRequest, *registry)

	expectedFacts := []entities.Fact{
		{
			Name:    "dummy1",
			Value:   &entities.FactValueInt{Value: 1},
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
	suite.Equal(suite.executionID, factResults.ExecutionID)
	suite.Equal(suite.agentID, factResults.AgentID)
	suite.Equal(suite.groupID, factResults.GroupID)
	suite.ElementsMatch(expectedFacts, factResults.FactsGathered)
}

func (suite *GatheringTestSuite) TestParentContextIsNotCancelledWhenGatherFails() {
	factsRequest := entities.FactsGatheringRequestedTarget{
		AgentID: suite.agentID,
		FactRequests: []entities.FactRequest{
			{
				Name:     "dummy1",
				Gatherer: "dummyGatherer1",
				Argument: "dummy1",
				CheckID:  "check1",
			},
		},
	}

	dummyGathererOne := &mocks.MockFactGatherer{}
	dummyGathererOne.
		On("Gather", mock.Anything, mock.Anything).
		Return(nil, errors.New("Gatherer error"))

	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		"dummyGatherer1": map[string]gatherers.FactGatherer{
			"v1": dummyGathererOne,
		},
	})

	ctx := context.Background()

	_, _ = gatherFacts(ctx, suite.executionID, suite.agentID, suite.groupID, &factsRequest, *registry)

	select {
	case <-ctx.Done():
		suite.Fail("Parent context should not be cancelled")
	default:
		suite.T().Log("Parent context is not cancelled")
	}
}

func (suite *GatheringTestSuite) TestGatherIsCancelledWhenParentContextIsCancelled() {
	factsRequest := entities.FactsGatheringRequestedTarget{
		AgentID: suite.agentID,
		FactRequests: []entities.FactRequest{
			{
				Name:     "dummy1",
				Gatherer: "dummyGatherer1",
				Argument: "dummy1",
				CheckID:  "check1",
			},
		},
	}

	dummyGathererOne := &mocks.MockFactGatherer{}
	dummyGathererOne.
		On("Gather", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			// nolint:forcetypeassert
			innerCtx := args.Get(0).(context.Context)
			select {
			case <-innerCtx.Done():
				suite.T().Log("Gather receives context cancellation")
			case <-time.After(3 * time.Second):
				suite.Fail("Gather should receive context cancellation")
			}
		}).
		Return([]entities.Fact{
			{
				Name:    "dummy1",
				Value:   &entities.FactValueInt{Value: 1},
				CheckID: "check1",
			},
		}, nil)

	registry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		"dummyGatherer1": map[string]gatherers.FactGatherer{
			"v1": dummyGathererOne,
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		cancel()
	}()
	_, err := gatherFacts(ctx, suite.executionID, suite.agentID, suite.groupID, &factsRequest, *registry)

	<-ctx.Done()

	if err == nil {
		suite.Fail("Error should not be nil")
	}

	suite.Equal(context.Canceled, err)

}
