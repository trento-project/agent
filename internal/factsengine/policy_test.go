// nolint:nosnakecase
package factsengine

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/adapters/mocks"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/contracts/pkg/events"
)

type PolicyTestSuite struct {
	suite.Suite
	mockAdapter mocks.Adapter
	factsEngine FactsEngine
}

func (suite *PolicyTestSuite) SetupTest() {
	suite.mockAdapter = mocks.Adapter{} // nolint
	suite.factsEngine = FactsEngine{    // nolint
		agentID:             agentID,
		factsServiceAdapter: &suite.mockAdapter,
		factGatherers:       map[string]gatherers.FactGatherer{},
	}
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) TestPolicyHandleEventWrongMessage() {
	err := suite.factsEngine.handleEvent("", []byte(""))
	suite.ErrorContains(err, "Error getting event type")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventInvalideEvent() {
	event, err := events.ToEvent(&events.FactsGathered{}, "", "") // nolint
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
	event, err := events.ToEvent(factsGatheringRequestsEvent, "", "") // nolint
	suite.NoError(err)

	err = suite.factsEngine.handleEvent("", event)
	suite.NoError(err)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEvent() {
	factsGatheringRequestsEvent := &events.FactsGatheringRequested{ // nolint
		Targets: []*events.FactsGatheringRequestedTarget{
			{
				AgentId: agentID,
			},
			{
				AgentId: "some-other-agent",
			},
		},
	}
	event, err := events.ToEvent(factsGatheringRequestsEvent, "", "") // nolint
	suite.NoError(err)

	suite.mockAdapter.On(
		"Publish",
		exchange,
		executionsRoutingKey,
		"",
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
		"",
		mock.MatchedBy(func(body []byte) bool {
			var facts events.FactsGathered
			err := events.FromEvent(body, &facts)
			if err != nil {
				panic(err)
			}

			expectedFacts := events.FactsGathered{
				AgentId:     agentID,
				ExecutionId: executionID,
				FactsGathered: []*events.Fact{
					{
						Name: "dummy1",
						Value: &events.Fact_TextValue{
							TextValue: "result1",
						},
						CheckId: "check1",
					},
					{
						Name: "dummy2",
						Value: &events.Fact_TextValue{
							TextValue: "result2",
						},
						CheckId: "check1",
					},
				},
			}

			suite.Equal(expectedFacts.AgentId, facts.AgentId)
			suite.Equal(expectedFacts.ExecutionId, facts.ExecutionId)
			suite.Equal(expectedFacts.FactsGathered, facts.FactsGathered)

			return true
		})).Return(nil)

	gatheredFacts := entities.FactsGathered{
		ExecutionID: executionID,
		AgentID:     agentID,
		FactsGathered: []entities.Fact{
			{
				Name:    "dummy1",
				Value:   "result1",
				CheckID: "check1",
			},
			{
				Name:    "dummy2",
				Value:   "result2",
				CheckID: "check1",
			},
		},
	}

	err := suite.factsEngine.publishFacts(gatheredFacts)

	suite.NoError(err)
}
