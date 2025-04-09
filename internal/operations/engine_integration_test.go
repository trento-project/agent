package operations_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/messaging"
	"github.com/trento-project/contracts/go/pkg/events"
	"github.com/trento-project/workbench/pkg/operator"
	operatorMocks "github.com/trento-project/workbench/pkg/operator/mocks"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/trento-project/agent/internal/operations"
)

type OperationsIntegrationTestSuite struct {
	suite.Suite
	amqpService     string
	rabbitmqAdapter messaging.Adapter
}

func TestFactsEngineIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(OperationsIntegrationTestSuite))
}

func (suite *OperationsIntegrationTestSuite) SetupSuite() {
	amqpService := os.Getenv("RABBITMQ_URL")
	if amqpService == "" {
		amqpService = "amqp://guest:guest@localhost:5675"
	}

	suite.amqpService = amqpService
}

func (suite *OperationsIntegrationTestSuite) SetupTest() {
	rabbitmqAdapter, err := messaging.NewRabbitMQAdapter(
		suite.amqpService,
		"trento.operations.requests",
		"trento.operations",
		"requests",
	)
	if err != nil {
		panic(err)
	}

	suite.rabbitmqAdapter = rabbitmqAdapter
}

func (suite *OperationsIntegrationTestSuite) TearDownTest() {
	if suite.rabbitmqAdapter == nil {
		return
	}

	err := suite.rabbitmqAdapter.Unsubscribe()
	if err != nil {
		panic(err)
	}
}

// nolint:nosnakecase
func (suite *OperationsIntegrationTestSuite) TestFactsEngineIntegration() {
	agentID := "some-agent"

	mockOperator := operatorMocks.NewMockOperator(suite.T())
	testRegistry := operator.NewRegistry(operator.OperatorBuildersTree{
		"test": map[string]operator.OperatorBuilder{
			"v1": func(_ string, _ operator.OperatorArguments) operator.Operator {
				return mockOperator
			},
		},
	})

	engine := operations.NewOperationsEngine(agentID, suite.amqpService, *testRegistry)

	err := engine.Subscribe()
	if err != nil {
		panic(err)
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	g, groupCtx := errgroup.WithContext(ctx)

	mockOperator.On(
		"Run",
		groupCtx,
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

	g.Go(func() error {
		return engine.Listen(groupCtx)
	})

	operatorExecutionRequested := events.OperatorExecutionRequested{
		OperationId: "some-operation",
		Operator:    "test@v1",
		Targets: []*events.OperatorExecutionRequestedTarget{
			{
				AgentId: agentID,
			},
		},
	}
	event, err := events.ToEvent(&operatorExecutionRequested, events.WithSource(""),
		events.WithID(""))
	if err != nil {
		panic(err)
	}

	handle := func(_ string, message []byte) error {
		defer ctxCancel()

		result := &events.OperatorExecutionCompleted_Value{
			Value: &events.OperatorResponse{
				Phase: events.OperatorPhase(events.OperatorPhase_value[string(operator.COMMIT)]),
				Diff: &events.OperatorDiff{
					Before: structpb.NewStringValue("before"),
					After:  structpb.NewStringValue("after"),
				},
			},
		}

		var operationCompleted events.OperatorExecutionCompleted
		err := events.FromEvent(message, &operationCompleted)
		suite.NoError(err)
		suite.Equal(agentID, operationCompleted.AgentId)
		suite.Equal("some-operation", operationCompleted.OperationId)
		suite.Equal(result, operationCompleted.Result)

		return nil
	}

	err = suite.rabbitmqAdapter.Listen(handle)
	if err != nil {
		panic(err)
	}

	time.Sleep(100 * time.Millisecond)

	err = suite.rabbitmqAdapter.Publish("agents", "", event)
	if err != nil {
		panic(err)
	}

	err = g.Wait()
	if err != nil {
		panic(err)
	}
}
