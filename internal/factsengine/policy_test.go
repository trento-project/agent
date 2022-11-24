// nolint:nosnakecase
package factsengine

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/adapters/mocks"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/contracts/go/pkg/events"
	"google.golang.org/protobuf/types/known/structpb"
)

type PolicyTestSuite struct {
	suite.Suite
	mockAdapter mocks.Adapter
	factsEngine FactsEngine
	executionID string
	agentID     string
	groupID     string
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) SetupTest() {
	suite.executionID = uuid.New().String()
	suite.agentID = uuid.New().String()
	suite.groupID = uuid.New().String()
	suite.mockAdapter = mocks.Adapter{} // nolint
	suite.factsEngine = FactsEngine{    // nolint
		agentID:             suite.agentID,
		factsServiceAdapter: &suite.mockAdapter,
	}
}

func (suite *PolicyTestSuite) TestPolicyHandleEventWrongMessage() {
	err := suite.factsEngine.handleEvent("", []byte(""))
	suite.ErrorContains(err, "Error getting event type")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventInvalideEvent() {
	event, err := events.ToEvent(
		&events.FactsGathered{}, // nolint
		events.WithSource(""),
		events.WithID(""),
	)
	suite.NoError(err)

	err = suite.factsEngine.handleEvent("", event)
	suite.EqualError(err, "Invalid event type: Trento.Checks.V1.FactsGathered")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventDiscardAgent() {
	factsGatheringRequestsEvent := &events.FactsGatheringRequested{ // nolint
		Targets: []*events.FactsGatheringRequestedTarget{
			{
				AgentId: "other-agent",
			},
			{
				AgentId: "some-other-agent",
			},
		},
	}
	event, err := events.ToEvent(
		factsGatheringRequestsEvent,
		events.WithSource(""),
		events.WithID(""),
	) // nolint
	suite.NoError(err)

	err = suite.factsEngine.handleEvent("", event)
	suite.NoError(err)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEvent() {
	factsGatheringRequestsEvent := &events.FactsGatheringRequested{ // nolint
		Targets: []*events.FactsGatheringRequestedTarget{
			{
				AgentId: suite.agentID,
			},
			{
				AgentId: "some-other-agent",
			},
		},
	}
	event, err := events.ToEvent(factsGatheringRequestsEvent, events.WithSource(""),
		events.WithID("")) // nolint
	suite.NoError(err)

	suite.mockAdapter.On(
		"Publish",
		exchange,
		executionsRoutingKey,
		events.ContentType(),
		mock.Anything).Return(nil)

	err = suite.factsEngine.handleEvent("", event)
	suite.NoError(err)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 1)
}

func (suite *PolicyTestSuite) TestPolicyPublishFacts() {
	suite.mockAdapter.On(
		"Publish",
		exchange,
		executionsRoutingKey,
		events.ContentType(),
		mock.MatchedBy(func(body []byte) bool {
			var facts events.FactsGathered
			err := events.FromEvent(body, &facts)
			if err != nil {
				panic(err)
			}

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
									StringValue: "result1",
								},
							},
						},
						CheckId: "check1",
					},
					{
						Name: "dummy2",
						FactValue: &events.Fact_Value{
							Value: &structpb.Value{
								Kind: &structpb.Value_StringValue{
									StringValue: "result2",
								},
							},
						},
						CheckId: "check1",
					},
				},
			}

			suite.Equal(expectedFacts.AgentId, facts.AgentId)
			suite.Equal(expectedFacts.ExecutionId, facts.ExecutionId)
			suite.Equal(expectedFacts.GroupId, facts.GroupId)
			suite.Equal(expectedFacts.FactsGathered, facts.FactsGathered)

			return true
		})).Return(nil)

	gatheredFacts := entities.FactsGathered{
		ExecutionID: suite.executionID,
		AgentID:     suite.agentID,
		GroupID:     suite.groupID,
		FactsGathered: []entities.Fact{
			{
				Name:    "dummy1",
				Value:   &entities.FactValueString{Value: "result1"},
				CheckID: "check1",
			},
			{
				Name:    "dummy2",
				Value:   &entities.FactValueString{Value: "result2"},
				CheckID: "check1",
			},
		},
	}

	err := suite.factsEngine.publishFacts(gatheredFacts)

	suite.NoError(err)
}
