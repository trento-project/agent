package operations_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/trento-project/agent/internal/operations"
	"github.com/trento-project/agent/internal/operations/operator"
	"github.com/trento-project/contracts/go/pkg/events"
)

type MapperTestSuite struct {
	suite.Suite
	operationID string
	groupID     string
	agentID     string
	stepNumber  int32
	operator    string
}

func TestMapperTestSuite(t *testing.T) {
	suite.Run(t, new(MapperTestSuite))
}

func (suite *MapperTestSuite) SetupSuite() {
	suite.operationID = uuid.New().String()
	suite.groupID = uuid.New().String()
	suite.agentID = uuid.New().String()
	suite.stepNumber = 1
	suite.operator = "some-operator"
}

func (suite *MapperTestSuite) TestOperatorExecutionRequestedFromEvent() {
	operatorExecution := events.OperatorExecutionRequested{
		OperationId: suite.operationID,
		GroupId:     suite.groupID,
		StepNumber:  suite.stepNumber,
		Operator:    suite.operator,
		Targets: []*events.OperatorExecutionRequestedTarget{
			{
				AgentId: "agent1",
				Arguments: map[string]*structpb.Value{
					"string": structpb.NewStringValue("foo"),
					"number": structpb.NewNumberValue(5),
				},
			},
			{
				AgentId: "agent2",
				Arguments: map[string]*structpb.Value{
					"other_string": structpb.NewStringValue("bar"),
					"other_number": structpb.NewNumberValue(10),
				},
			},
		},
	}

	eventBytes, err := events.ToEvent(
		&operatorExecution,
		events.WithSource("source"),
		events.WithID("id"),
	)
	suite.NoError(err)

	request, err := operations.OperatorExecutionRequestedFromEvent(eventBytes)
	suite.NoError(err)

	expectedRequest := &operations.OperatorExecutionRequested{
		OperationID: suite.operationID,
		GroupID:     suite.groupID,
		StepNumber:  suite.stepNumber,
		Operator:    suite.operator,
		Targets: []operations.OperatorExecutionRequestedTarget{
			{
				AgentID: "agent1",
				Arguments: map[string]interface{}{
					"string": "foo",
					"number": 5.0,
				},
			},
			{
				AgentID: "agent2",
				Arguments: map[string]interface{}{
					"other_string": "bar",
					"other_number": 10.0,
				},
			},
		},
	}

	suite.Equal(expectedRequest, request)
}

func (suite *MapperTestSuite) TestOperatorExecutionRequestedFromEventError() {
	_, err := operations.OperatorExecutionRequestedFromEvent([]byte("error"))
	suite.Error(err)
}

func (suite *MapperTestSuite) TestGetTargetAgentFound() {
	request := &operations.OperatorExecutionRequested{
		Targets: []operations.OperatorExecutionRequestedTarget{
			{
				AgentID: "agent1",
			},
			{
				AgentID: "agent2",
			},
		},
	}

	suite.Equal("agent2", request.GetTargetAgent("agent2").AgentID)
}

func (suite *MapperTestSuite) TestGetTargetAgentNotFound() {
	request := &operations.OperatorExecutionRequested{
		Targets: []operations.OperatorExecutionRequestedTarget{
			{
				AgentID: "agent1",
			},
			{
				AgentID: "agent2",
			},
		},
	}

	suite.Nil(request.GetTargetAgent("agent3"))
}

func (suite *MapperTestSuite) TestOperatorExecutionCompletedToEventSuccess() {
	event, err := operations.OperatorExecutionCompletedToEvent(
		suite.operationID,
		suite.groupID,
		suite.agentID,
		suite.stepNumber,
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

	suite.NoError(err)

	var operation events.OperatorExecutionCompleted
	err = events.FromEvent(event, &operation)
	suite.NoError(err)

	expectedResult := &events.OperatorExecutionCompleted_Value{
		Value: &events.OperatorResponse{
			Phase: events.OperatorPhase(events.OperatorPhase_value[string(operator.COMMIT)]),
			Diff: &events.OperatorDiff{
				Before: structpb.NewStringValue("before"),
				After:  structpb.NewStringValue("after"),
			},
		},
	}

	suite.Equal(suite.operationID, operation.OperationId)
	suite.Equal(suite.groupID, operation.GroupId)
	suite.Equal(suite.agentID, operation.AgentId)
	suite.Equal(suite.stepNumber, operation.StepNumber)
	suite.Equal(expectedResult, operation.Result)
}

func (suite *MapperTestSuite) TestOperatorExecutionCompletedToEventSuccessBeforeMissing() {
	_, err := operations.OperatorExecutionCompletedToEvent(
		suite.operationID,
		suite.groupID,
		suite.agentID,
		suite.stepNumber,
		&operator.ExecutionReport{
			Success: &operator.ExecutionSuccess{
				Diff: map[string]any{
					"after": "after",
				},
				LastPhase: operator.COMMIT,
			},
		},
	)

	suite.ErrorContains(err, "before not found in report")
}

func (suite *MapperTestSuite) TestOperatorExecutionCompletedToEventSuccessAfterMissing() {
	_, err := operations.OperatorExecutionCompletedToEvent(
		suite.operationID,
		suite.groupID,
		suite.agentID,
		suite.stepNumber,
		&operator.ExecutionReport{
			Success: &operator.ExecutionSuccess{
				Diff: map[string]any{
					"before": "before",
				},
				LastPhase: operator.COMMIT,
			},
		},
	)

	suite.ErrorContains(err, "after not found in report")
}

func (suite *MapperTestSuite) TestOperatorExecutionCompletedToEventFailure() {
	event, err := operations.OperatorExecutionCompletedToEvent(
		suite.operationID,
		suite.groupID,
		suite.agentID,
		suite.stepNumber,
		&operator.ExecutionReport{
			Error: &operator.ExecutionError{
				Message:    "error message",
				ErrorPhase: operator.COMMIT,
			},
		},
	)

	suite.NoError(err)

	var operation events.OperatorExecutionCompleted
	err = events.FromEvent(event, &operation)
	suite.NoError(err)

	expectedResult := &events.OperatorExecutionCompleted_Error{
		Error: &events.OperatorError{
			Phase:   events.OperatorPhase(events.OperatorPhase_value[string(operator.COMMIT)]),
			Message: "error message",
		},
	}

	suite.Equal(suite.operationID, operation.OperationId)
	suite.Equal(suite.groupID, operation.GroupId)
	suite.Equal(suite.agentID, operation.AgentId)
	suite.Equal(suite.stepNumber, operation.StepNumber)
	suite.Equal(expectedResult, operation.Result)
}
