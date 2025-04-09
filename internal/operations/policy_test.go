package operations_test

import (
	"context"
	"testing"

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
	suite.ErrorContains(err, "Error getting event type")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventInvalideEvent() {
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
	suite.EqualError(err, "Invalid event type: Trento.Operations.V1.OperatorExecutionCompleted")
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
