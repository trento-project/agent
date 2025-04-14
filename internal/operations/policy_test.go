package operations_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/messaging/mocks"
	"github.com/trento-project/contracts/go/pkg/events"
	"github.com/trento-project/workbench/pkg/operator"
	operatorMocks "github.com/trento-project/workbench/pkg/operator/mocks"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/trento-project/agent/internal/operations"
)

type PolicyTestSuite struct {
	suite.Suite
	agentID      string
	mockAdapter  mocks.Adapter
	mockOperator *operatorMocks.MockOperator
	testRegistry *operator.Registry
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) SetupTest() {
	suite.agentID = uuid.New().String()
	suite.mockAdapter = mocks.Adapter{} // nolint
	suite.mockOperator = operatorMocks.NewMockOperator(suite.T())
	suite.testRegistry = operator.NewRegistry(operator.OperatorBuildersTree{
		"test": map[string]operator.OperatorBuilder{
			"v1": func(_ string, _ operator.OperatorArguments) operator.Operator {
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
	suite.ErrorContains(err, "error getting event type")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventInvalidEvent() {
	event, err := events.ToEvent(
		&events.OperatorExecutionCompleted{}, // nolint
		events.WithSource(""),
		events.WithID(""),
	)
	suite.NoError(err)

	err = operations.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.EqualError(err, "invalid event type: Trento.Operations.V1.OperatorExecutionCompleted")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventDiscardAgent() {
	operatorRequestsEvent := &events.OperatorExecutionRequested{ // nolint
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
	) // nolint
	suite.NoError(err)

	err = operations.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)
	suite.NoError(err)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventOperatorNotFound() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{ // nolint
		Operator: "foo",
		Targets: []*events.OperatorExecutionRequestedTarget{
			{
				AgentId: suite.agentID,
			},
		},
	}
	event, err := events.ToEvent(operatorRequestsEvent,
		events.WithSource(""),
		events.WithID("")) // nolint
	suite.NoError(err)

	err = operations.HandleEvent(
		ctx,
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)

	suite.EqualError(err, "error building operator from operators registry: operator foo not found")
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventErrorDecoding() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{} // nolint

	now := time.Now()
	expiration := now.Add(-60 * time.Minute)
	event, err := events.ToEvent(operatorRequestsEvent,
		events.WithTime(now),
		events.WithExpiration(expiration),
		events.WithSource(""),
		events.WithID("")) // nolint
	suite.NoError(err)

	err = operations.HandleEvent(
		ctx,
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)

	suite.EqualError(err, "error decoding OperatorExecutionRequested event: "+
		"cannot decode cloudevent, event expired")
	suite.mockOperator.AssertNumberOfCalls(suite.T(), "Run", 0)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventErrorEncoding() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{ // nolint
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
		events.WithID("")) // nolint
	suite.NoError(err)

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

	suite.EqualError(err, "error encoding OperatorExecutionCompleted event: after not found in report")
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventErrorPublishing() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{ // nolint
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
		events.WithID("")) // nolint
	suite.NoError(err)

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
	).Return(fmt.Errorf("publishing error"))

	err = operations.HandleEvent(
		ctx,
		event,
		suite.agentID,
		&suite.mockAdapter,
		*suite.testRegistry,
	)

	suite.EqualError(err, "error publishing operator execution report: publishing error")
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 1)
}

func (suite *PolicyTestSuite) TestPolicyHandleEvent() {
	ctx := context.Background()

	operatorRequestsEvent := &events.OperatorExecutionRequested{ // nolint
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
		events.WithID("")) // nolint
	suite.NoError(err)

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
	suite.NoError(err)
	suite.mockAdapter.AssertNumberOfCalls(suite.T(), "Publish", 1)
}
