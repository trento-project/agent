// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package factsengine_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/v3/internal/factsengine/gatherers"
	gathererMocks "github.com/trento-project/agent/v3/internal/factsengine/gatherers/mocks"
	"github.com/trento-project/agent/v3/internal/messaging/mocks"
	"github.com/trento-project/agent/v3/pkg/factsengine/entities"
	"github.com/trento-project/contracts/go/pkg/events"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/trento-project/agent/v3/internal/factsengine"
)

type PolicyTestSuite struct {
	suite.Suite

	mockAdapter  mocks.MockAdapter
	executionID  string
	agentID      string
	groupID      string
	mockGatherer gathererMocks.MockFactGatherer
	testRegistry *gatherers.Registry
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) SetupTest() {
	suite.executionID = uuid.New().String()
	suite.agentID = uuid.New().String()
	suite.groupID = uuid.New().String()
	suite.mockAdapter = mocks.MockAdapter{}
	suite.mockGatherer = gathererMocks.MockFactGatherer{}
	suite.testRegistry = gatherers.NewRegistry(gatherers.FactGatherersTree{
		"test": map[string]gatherers.FactGatherer{
			"v1": &suite.mockGatherer,
		},
	})
}

func (suite *PolicyTestSuite) TestPolicyHandleEventWrongMessage() {
	err := factsengine.HandleEvent(
		context.Background(),
		[]byte(""),
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.Require().ErrorContains(err, "Error getting event type")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventInvalidEvent() {
	event, err := events.ToEvent(
		&events.FactsGathered{},
		events.WithSource(""),
		events.WithID(""),
	)
	suite.Require().NoError(err)

	err = factsengine.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.Require().EqualError(err, "Invalid event type: Trento.Checks.V1.FactsGathered")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventDiscardAgent() {
	factsGatheringRequestsEvent := &events.FactsGatheringRequested{
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
	)
	suite.Require().NoError(err)

	err = factsengine.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.Require().NoError(err)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEvent() {
	factsGatheringRequestsEvent := &events.FactsGatheringRequested{
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
		events.WithID(""))
	suite.Require().NoError(err)

	suite.mockAdapter.On(
		"Publish",
		"executions",
		events.ContentType(),
		mock.Anything,
	).Return(nil)

	err = factsengine.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.Require().NoError(err)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 1)
}

func (suite *PolicyTestSuite) TestPolicyPublishFacts() {
	ctx := context.Background()
	factsGatheringRequestsEvent := &events.FactsGatheringRequested{
		ExecutionId: suite.executionID,
		GroupId:     suite.groupID,
		Targets: []*events.FactsGatheringRequestedTarget{
			{
				AgentId: suite.agentID,
				FactRequests: []*events.FactRequest{
					{
						Gatherer: "test",
					},
				},
			},
		},
	}
	event, err := events.ToEvent(factsGatheringRequestsEvent, events.WithSource(""),
		events.WithID(""))
	suite.Require().NoError(err)

	suite.mockAdapter.On(
		"Publish",
		"executions",
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

			suite.Equal(expectedFacts.GetAgentId(), facts.GetAgentId())
			suite.Equal(expectedFacts.GetExecutionId(), facts.GetExecutionId())
			suite.Equal(expectedFacts.GetGroupId(), facts.GetGroupId())
			suite.Equal(expectedFacts.GetFactsGathered(), facts.GetFactsGathered())

			return true
		})).Return(nil)

	suite.mockGatherer.
		On(
			"Gather",
			ctx,
			mock.Anything).
		Return(
			[]entities.Fact{
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
			nil,
		)

	err = factsengine.HandleEvent(
		ctx,
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)

	suite.Require().NoError(err)
}
