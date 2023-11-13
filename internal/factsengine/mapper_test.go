// nolint:nosnakecase
package factsengine_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/trento-project/agent/internal/factsengine"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/contracts/go/pkg/events"
)

type MapperTestSuite struct {
	suite.Suite
	executionID string
	agentID     string
	groupID     string
}

func TestMapperTestSuite(t *testing.T) {
	suite.Run(t, new(MapperTestSuite))
}

func (suite *MapperTestSuite) SetupSuite() {
	suite.executionID = uuid.New().String()
	suite.agentID = uuid.New().String()
	suite.groupID = uuid.New().String()
}

func (suite *MapperTestSuite) TestFactsGatheredToEvent() {
	factsGathered := entities.FactsGathered{
		ExecutionID: suite.executionID,
		AgentID:     suite.agentID,
		GroupID:     suite.groupID,
		FactsGathered: []entities.Fact{
			{
				Name:    "dummy1",
				Value:   &entities.FactValueString{Value: "result"},
				CheckID: "check1",
			},
			{
				Name:    "dummy2",
				Value:   &entities.FactValueInt{Value: 2},
				CheckID: "check1",
			},
			{
				Name:    "dummy3",
				Value:   &entities.FactValueFloat{Value: 2.0},
				CheckID: "check1",
			},
			{
				Name:    "dummy4",
				Value:   &entities.FactValueBool{Value: true},
				CheckID: "check1",
			},
			{
				Name: "dummy5",
				Value: &entities.FactValueList{
					Value: []entities.FactValue{
						&entities.FactValueString{Value: "a"},
						&entities.FactValueString{Value: "b"},
					},
				},
				CheckID: "check1",
			},
			{
				Name: "dummy5",
				Value: &entities.FactValueMap{
					Value: map[string]entities.FactValue{
						"a": &entities.FactValueString{Value: "c"},
					},
				},
				CheckID: "check1",
			},
		},
	}

	result, err := factsengine.FactsGatheredToEvent(factsGathered)
	suite.NoError(err)

	var facts events.FactsGathered
	err = events.FromEvent(result, &facts)
	suite.NoError(err)

	expectedFacts := events.FactsGathered{
		AgentId:     suite.agentID,
		ExecutionId: suite.executionID,
		GroupId:     suite.groupID,
		FactsGathered: []*events.Fact{
			{
				Name: "dummy1",
				FactValue: &events.Fact_Value{
					Value: &structpb.Value{
						Kind: &structpb.Value_StringValue{
							StringValue: "result",
						},
					},
				},
				CheckId: "check1",
			},
			{
				Name: "dummy2",
				FactValue: &events.Fact_Value{
					Value: &structpb.Value{
						Kind: &structpb.Value_NumberValue{
							NumberValue: float64(2),
						},
					},
				},
				CheckId: "check1",
			},
			{
				Name: "dummy3",
				FactValue: &events.Fact_Value{
					Value: &structpb.Value{
						Kind: &structpb.Value_NumberValue{
							NumberValue: 2.0,
						},
					},
				},
				CheckId: "check1",
			},
			{
				Name: "dummy4",
				FactValue: &events.Fact_Value{
					Value: &structpb.Value{
						Kind: &structpb.Value_BoolValue{
							BoolValue: true,
						},
					},
				},
				CheckId: "check1",
			},
			{
				Name: "dummy5",
				FactValue: &events.Fact_Value{
					Value: &structpb.Value{
						Kind: &structpb.Value_ListValue{
							ListValue: &structpb.ListValue{
								Values: []*structpb.Value{
									{
										Kind: &structpb.Value_StringValue{
											StringValue: "a",
										},
									},
									{
										Kind: &structpb.Value_StringValue{
											StringValue: "b",
										},
									},
								},
							},
						},
					},
				},
				CheckId: "check1",
			},
			{
				Name: "dummy5",
				FactValue: &events.Fact_Value{
					Value: &structpb.Value{
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"a": {
										Kind: &structpb.Value_StringValue{StringValue: "c"},
									},
								},
							},
						},
					},
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
	factsGathered := entities.FactsGathered{
		ExecutionID: suite.executionID,
		AgentID:     suite.agentID,
		GroupID:     suite.groupID,
		FactsGathered: []entities.Fact{
			{
				Name:    "dummy1",
				Value:   nil,
				CheckID: "check1",
				Error: &entities.FactGatheringError{
					Message: "some message",
					Type:    "some_type",
				},
			},
			{
				Name:    "dummy2",
				Value:   &entities.FactValueString{Value: "result"},
				CheckID: "check1",
			},
		},
	}

	result, err := factsengine.FactsGatheredToEvent(factsGathered)
	suite.NoError(err)

	var facts events.FactsGathered
	err = events.FromEvent(result, &facts)
	suite.NoError(err)

	expectedFacts := events.FactsGathered{
		AgentId:     suite.agentID,
		ExecutionId: suite.executionID,
		GroupId:     suite.groupID,
		FactsGathered: []*events.Fact{
			{
				Name: "dummy1",
				FactValue: &events.Fact_ErrorValue{
					ErrorValue: &events.FactError{
						Message: "some message",
						Type:    "some_type",
					},
				},
				CheckId: "check1",
			},
			{
				Name: "dummy2",
				FactValue: &events.Fact_Value{
					Value: &structpb.Value{
						Kind: &structpb.Value_StringValue{
							StringValue: "result",
						},
					},
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

	eventBytes, err := events.ToEvent(
		&event,
		events.WithSource("source"),
		events.WithID("id"),
	)
	suite.NoError(err)

	request, err := factsengine.FactsGatheringRequestedFromEvent(eventBytes)
	expectedRequest := &entities.FactsGatheringRequested{
		ExecutionID: "executionID",
		GroupID:     "groupID",
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
	_, err := factsengine.FactsGatheringRequestedFromEvent([]byte("error"))
	suite.Error(err)
}
