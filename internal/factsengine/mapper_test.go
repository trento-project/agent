// nolint:nosnakecase
package factsengine

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/contracts/pkg/events"
)

type MapperTestSuite struct {
	suite.Suite
}

func TestMapperTestSuite(t *testing.T) {
	suite.Run(t, new(MapperTestSuite))
}

func (suite *MapperTestSuite) TestFactsGatheredToEvent() {
	someID := uuid.New().String()
	someAgent := uuid.New().String()

	factsGathered := entities.FactsGathered{
		ExecutionID: someID,
		AgentID:     someAgent,
		FactsGathered: []entities.Fact{
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
		},
	}

	result, err := FactsGatheredToEvent(factsGathered)
	suite.NoError(err)

	var facts events.FactsGathered
	err = events.FromEvent(result, &facts)
	suite.NoError(err)

	expectedFacts := events.FactsGathered{
		AgentId:     someAgent,
		ExecutionId: someID,
		FactsGathered: []*events.Fact{
			{
				Name: "dummy1",
				Value: &events.Fact_TextValue{
					TextValue: "1",
				},
				CheckId: "check1",
			},
			{
				Name: "dummy2",
				Value: &events.Fact_TextValue{
					TextValue: "2",
				},
				CheckId: "check1",
			},
		},
	}

	suite.Equal(expectedFacts.AgentId, facts.AgentId)
	suite.Equal(expectedFacts.ExecutionId, facts.ExecutionId)
	suite.Equal(expectedFacts.FactsGathered, facts.FactsGathered)
}

func (suite *MapperTestSuite) TestFactsGatheredWithErrorToEvent() {
	someID := uuid.New().String()
	someAgent := uuid.New().String()

	factsGathered := entities.FactsGathered{
		ExecutionID: someID,
		AgentID:     someAgent,
		FactsGathered: []entities.Fact{
			{
				Name:    "dummy1",
				Value:   nil,
				CheckID: "check1",
				Error: &entities.Error{
					Message: "some message",
					Type:    "some_type",
				},
			},
			{
				Name:    "dummy2",
				Value:   "2",
				CheckID: "check1",
			},
		},
	}

	result, err := FactsGatheredToEvent(factsGathered)
	suite.NoError(err)

	var facts events.FactsGathered
	err = events.FromEvent(result, &facts)
	suite.NoError(err)

	expectedFacts := events.FactsGathered{
		AgentId:     someAgent,
		ExecutionId: someID,
		FactsGathered: []*events.Fact{
			{
				Name: "dummy1",
				Value: &events.Fact_ErrorValue{
					ErrorValue: &events.FactError{
						Message: "some message",
						Type:    "some_type",
					},
				},
				CheckId: "check1",
			},
			{
				Name: "dummy2",
				Value: &events.Fact_TextValue{
					TextValue: "2",
				},
				CheckId: "check1",
			},
		},
	}

	suite.Equal(expectedFacts.AgentId, facts.AgentId)
	suite.Equal(expectedFacts.ExecutionId, facts.ExecutionId)
	suite.Equal(expectedFacts.FactsGathered, facts.FactsGathered)
}

func (suite *MapperTestSuite) TestFactsGatheringRequestedFromEvent() {

	event := events.FactsGatheringRequested{
		ExecutionId: "executionID",
		GroupId:     "groupID",
		Targets: []*events.FactsGatheringRequestedTarget{
			{
				AgentId: "agent1",
				FactRequests: []*events.FactRequest{
					{
						Argument: "argument1",
						CheckId:  "check1",
						Gatherer: "gatherer1",
						Name:     "name1",
					},
					{
						Argument: "argument2",
						CheckId:  "check2",
						Gatherer: "gatherer2",
						Name:     "name2",
					},
				},
			},
			{
				AgentId: "agent2",
				FactRequests: []*events.FactRequest{
					{
						Argument: "argument1",
						CheckId:  "check1",
						Gatherer: "gatherer1",
						Name:     "name1",
					},
					{
						Argument: "argument2",
						CheckId:  "check2",
						Gatherer: "gatherer2",
						Name:     "name2",
					},
				},
			},
		},
	}

	eventBytes, err := events.ToEvent(&event, "source", "id")
	suite.NoError(err)

	request, err := FactsGatheringRequestedFromEvent(eventBytes)
	expectedRequest := &entities.FactsGatheringRequested{
		ExecutionID: "executionID",
		Targets: []entities.FactsGatheringRequestedTarget{
			{
				AgentID: "agent1",
				FactRequests: []entities.FactRequest{
					{
						Argument: "argument1",
						CheckID:  "check1",
						Gatherer: "gatherer1",
						Name:     "name1",
					},
					{
						Argument: "argument2",
						CheckID:  "check2",
						Gatherer: "gatherer2",
						Name:     "name2",
					},
				},
			},
			{
				AgentID: "agent2",
				FactRequests: []entities.FactRequest{
					{
						Argument: "argument1",
						CheckID:  "check1",
						Gatherer: "gatherer1",
						Name:     "name1",
					},
					{
						Argument: "argument2",
						CheckID:  "check2",
						Gatherer: "gatherer2",
						Name:     "name2",
					},
				},
			},
		},
	}

	suite.NoError(err)
	suite.Equal(expectedRequest, request)
}

func (suite *MapperTestSuite) TestFactsGatheringRequestedFromEventError() {
	_, err := FactsGatheringRequestedFromEvent([]byte("error"))
	suite.Error(err)
}
