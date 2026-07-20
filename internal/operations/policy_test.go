// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package operations_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/messaging/mocks"
	"github.com/trento-project/agent/internal/operations/operator"
	operatorMocks "github.com/trento-project/agent/internal/operations/operator/mocks"
	"github.com/trento-project/contracts/go/pkg/events"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/trento-project/agent/internal/operations"
)

type PolicyTestSuite struct {
	suite.Suite

	agentID      string
	mockAdapter  mocks.MockAdapter
	mockOperator *operatorMocks.MockOperator
	testRegistry *operator.Registry
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) SetupTest() {
	suite.agentID = uuid.New().String()
	suite.mockAdapter = mocks.MockAdapter{}
	suite.mockOperator = operatorMocks.NewMockOperator(suite.T())
	suite.testRegistry = operator.NewRegistry(operator.BuildersTree{
		"test": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator {
				return suite.mockOperator
			},
		},
	})
}

func (suite *PolicyTestSuite) TestPolicyHandleEventWrongMessage() {
	err := operations.HandleEvent(
		context.Background(),
		[]byte(""),
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.Require().ErrorContains(err, "error getting event type")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventInvalidEvent() {
	event, err := events.ToEvent(
		&events.OperatorExecutionCompleted{},
		events.WithSource(""),
		events.WithID(""),
	)
	suite.Require().NoError(err)

	err = operations.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.Require().EqualError(err, "invalid event type: Trento.Operations.V1.OperatorExecutionCompleted")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventDiscardAgent() {
	operatorRequestsEvent := &events.OperatorExecutionRequested{
		Targets: []*events.OperatorExecutionRequestedTarget{
			{
				AgentId: "other-agent",
			},
			{
				AgentId: "some-other-agent",
			},
		},
	}
	event, err := events.ToEvent(
		operatorRequestsEvent,
		events.WithSource(""),
		events.WithID(""),
	)
	suite.Require().NoError(err)

	err = operations.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.Require().NoError(err)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventOperatorNotFound() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{
		Operator: "foo",
		Targets: []*events.OperatorExecutionRequestedTarget{
			{
				AgentId: suite.agentID,
			},
		},
	}
	event, err := events.ToEvent(operatorRequestsEvent,
		events.WithSource(""),
		events.WithID(""))
	suite.Require().NoError(err)

	err = operations.HandleEvent(
		ctx,
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)

	suite.Require().EqualError(err, "error building operator from operators registry: operator foo not found")
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventErrorDecoding() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{}

	now := time.Now()
	expiration := now.Add(-60 * time.Minute)
	event, err := events.ToEvent(operatorRequestsEvent,
		events.WithTime(now),
		events.WithExpiration(expiration),
		events.WithSource(""),
		events.WithID(""))
	suite.Require().NoError(err)

	err = operations.HandleEvent(
		ctx,
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)

	suite.Require().EqualError(err, "error decoding OperatorExecutionRequested event: "+
		"cannot decode cloudevent, event expired")
	suite.mockOperator.AssertNumberOfCalls(suite.T(), "Run", 0)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventErrorEncoding() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{
		OperationId: uuid.New().String(),
		Operator:    "test@v1",
		Targets: []*events.OperatorExecutionRequestedTarget{
			{
				AgentId:   suite.agentID,
				Arguments: map[string]*structpb.Value{},
			},
		},
	}
	event, err := events.ToEvent(operatorRequestsEvent,
		events.WithSource(""),
		events.WithID(""))
	suite.Require().NoError(err)

	suite.mockOperator.On(
		"Run",
		ctx,
	).Return(
		&operator.ExecutionReport{
			Success: &operator.ExecutionSuccess{
				Diff: map[string]any{
					"before": "before",
				},
				LastPhase: operator.COMMIT,
			},
		},
	)

	err = operations.HandleEvent(
		ctx,
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)

	suite.Require().EqualError(err, "error encoding OperatorExecutionCompleted event: after not found in report")
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventErrorPublishing() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{
		OperationId: uuid.New().String(),
		Operator:    "test@v1",
		Targets: []*events.OperatorExecutionRequestedTarget{
			{
				AgentId:   suite.agentID,
				Arguments: map[string]*structpb.Value{},
			},
		},
	}
	event, err := events.ToEvent(operatorRequestsEvent,
		events.WithSource(""),
		events.WithID(""))
	suite.Require().NoError(err)

	suite.mockOperator.On(
		"Run",
		ctx,
	).Return(
		&operator.ExecutionReport{
			Success: &operator.ExecutionSuccess{
				Diff: map[string]any{
					"before": "before",
					"after":  "after",
				},
				LastPhase: operator.COMMIT,
			},
		},
	)

	suite.mockAdapter.On(
		"Publish",
		"requests",
		events.ContentType(),
		mock.Anything,
	).Return(errors.New("publishing error"))

	err = operations.HandleEvent(
		ctx,
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)

	suite.Require().EqualError(err, "error publishing operator execution report: publishing error")
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 1)
}

func (suite *PolicyTestSuite) TestPolicyHandleEvent() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{
		OperationId: uuid.New().String(),
		Operator:    "test@v1",
		Targets: []*events.OperatorExecutionRequestedTarget{
			{
				AgentId:   suite.agentID,
				Arguments: map[string]*structpb.Value{},
			},
			{
				AgentId:   "some-other-agent",
				Arguments: map[string]*structpb.Value{},
			},
		},
	}
	event, err := events.ToEvent(operatorRequestsEvent,
		events.WithSource(""),
		events.WithID(""))
	suite.Require().NoError(err)

	suite.mockOperator.On(
		"Run",
		ctx,
	).Return(
		&operator.ExecutionReport{
			Success: &operator.ExecutionSuccess{
				Diff: map[string]any{
					"before": "before",
					"after":  "after",
				},
				LastPhase: operator.COMMIT,
			},
		},
	)

	suite.mockAdapter.On(
		"Publish",
		"requests",
		events.ContentType(),
		mock.Anything,
	).Return(nil)

	err = operations.HandleEvent(
		ctx,
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.Require().NoError(err)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 1)
}
