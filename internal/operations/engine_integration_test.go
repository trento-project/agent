// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package operations_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/messaging/testsupport"
	"github.com/trento-project/agent/internal/operations/operator"
	operatorMocks "github.com/trento-project/agent/internal/operations/operator/mocks"
	"github.com/trento-project/contracts/go/pkg/events"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/trento-project/agent/internal/operations"
)

type OperationsIntegrationTestSuite struct {
	testsupport.RabbitMQIntegrationSuite
}

func TestFactsEngineIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	s := &OperationsIntegrationTestSuite{
		RabbitMQIntegrationSuite: testsupport.RabbitMQIntegrationSuite{
			QueueName:    "trento.operations.requests",
			ExchangeName: "trento.operations",
			RoutingKey:   "requests",
		},
	}

	suite.Run(t, s)
}

func (suite *OperationsIntegrationTestSuite) TestFactsEngineIntegration() {
	agentID := "some-agent"

	mockOperator := operatorMocks.NewMockOperator(suite.T())
	testRegistry := operator.NewRegistry(operator.BuildersTree{
		"test": map[string]operator.Builder{
			"v1": func(_ string, _ operator.Arguments) operator.Operator {
				return mockOperator
			},
		},
	})

	engine := operations.NewOperationsEngine(agentID, suite.AMQPService, *testRegistry)

	err := engine.Subscribe()
	if err != nil {
		panic(err)
	}

	// Bounded as a safety net: if the request is never picked up, fail fast
	// with a clear error instead of hanging until the outer test timeout.
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()
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
		suite.Require().NoError(err)
		suite.Equal(agentID, operationCompleted.GetAgentId())
		suite.Equal("some-operation", operationCompleted.GetOperationId())
		suite.Equal(result, operationCompleted.GetResult())

		return nil
	}

	err = suite.RabbitMQAdapter.Listen(handle)
	if err != nil {
		panic(err)
	}

	testsupport.PublishUntilDone(groupCtx, suite.RabbitMQAdapter, "agents", event,
		"failed to publish operation request, will retry")

	err = g.Wait()
	if err != nil {
		panic(err)
	}
}
