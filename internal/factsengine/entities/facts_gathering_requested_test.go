// nolint:nosnakecase
package entities

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/trento-project/contracts/golang/pkg/events"
)

type FactsGatheringRequestedTestSuite struct {
	suite.Suite
}

func TestFactsGatheringRequestedTestSuite(t *testing.T) {
	suite.Run(t, new(FactsGatheringRequestedTestSuite))
}

func (suite *FactsGatheringRequestedTestSuite) TestFactsGatheringRequestedFromEvent() {

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
	expectedRequest := &FactsGatheringRequested{
		ExecutionID: "executionID",
		Targets: []FactsGatheringRequestedTarget{
			{
				AgentID: "agent1",
				FactRequests: []FactRequest{
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
				FactRequests: []FactRequest{
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

func (suite *FactsGatheringRequestedTestSuite) TestFactsGatheringRequestedFromEventError() {
	_, err := FactsGatheringRequestedFromEvent([]byte("error"))
	suite.EqualError(err, "proto: invalid nil source message")
}
