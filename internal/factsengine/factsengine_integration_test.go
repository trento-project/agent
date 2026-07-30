// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package factsengine_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/trento-project/agent/v3/internal/factsengine"
	"github.com/trento-project/agent/v3/internal/factsengine/gatherers"
	"github.com/trento-project/agent/v3/internal/messaging/testsupport"
	"github.com/trento-project/agent/v3/pkg/factsengine/entities"
	"github.com/trento-project/contracts/go/pkg/events"
)

type FactsEngineIntegrationTestSuite struct {
	testsupport.RabbitMQIntegrationSuite
}

func TestFactsEngineIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	s := &FactsEngineIntegrationTestSuite{
		RabbitMQIntegrationSuite: testsupport.RabbitMQIntegrationSuite{
			QueueName:    "trento.checks.executions",
			ExchangeName: "trento.checks",
			RoutingKey:   "executions",
		},
	}

	suite.Run(t, s)
}

type FactsEngineIntegrationTestGatherer struct{}

func NewFactsEngineIntegrationTestGatherer() *FactsEngineIntegrationTestGatherer {
	return &FactsEngineIntegrationTestGatherer{}
}

func (s *FactsEngineIntegrationTestGatherer) Gather(_ context.Context, requests []entities.FactRequest) ([]entities.Fact, error) {
	facts := make([]entities.Fact, 0, len(requests))

	for i, req := range requests {
		fact := entities.Fact{
			Name:    req.Name,
			Value:   &entities.FactValueInt{Value: i},
			CheckID: req.CheckID,
			Error:   nil,
		}
		facts = append(facts, fact)
	}

	return facts, nil
}

func (suite *FactsEngineIntegrationTestSuite) TestFactsEngineIntegration() {
	agentID := "some-agent"

	gathererRegistry := gatherers.NewRegistry(gatherers.FactGatherersTree{
		"integration": map[string]gatherers.FactGatherer{
			"v1": NewFactsEngineIntegrationTestGatherer(),
		},
	})

	engine := factsengine.NewFactsEngine(agentID, suite.AMQPService, *gathererRegistry)

	err := engine.Subscribe()
	if err != nil {
		panic(err)
	}

	// Bounded as a safety net: if the request is never picked up, fail fast
	// with a clear error instead of hanging until the outer test timeout.
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()
	g, groupCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return engine.Listen(groupCtx)
	})

	factGatheringRequested := events.FactsGatheringRequested{
		ExecutionId: "some-execution",
		GroupId:     "",
		Targets: []*events.FactsGatheringRequestedTarget{
			{
				AgentId: agentID,
				FactRequests: []*events.FactRequest{
					{
						CheckId:  "check1",
						Name:     "test1",
						Gatherer: "integration",
						Argument: "arg1",
					},
					{
						CheckId:  "check2",
						Name:     "test2",
						Gatherer: "integration",
						Argument: "arg2",
					},
				},
			},
		},
	}

	event, err := events.ToEvent(&factGatheringRequested, events.WithSource(""),
		events.WithID(""))
	if err != nil {
		panic(err)
	}

	handle := func(_ string, message []byte) error {
		defer ctxCancel()

		expectedFactsGathered := events.FactsGathered{
			AgentId:     agentID,
			ExecutionId: "some-execution",
			FactsGathered: []*events.Fact{
				{
					CheckId: "check1",
					Name:    "test1",
					FactValue: &events.Fact_Value{
						Value: &structpb.Value{
							Kind: &structpb.Value_NumberValue{
								NumberValue: float64(0),
							},
						},
					},
				},
				{
					CheckId: "check2",
					Name:    "test2",
					FactValue: &events.Fact_Value{
						Value: &structpb.Value{
							Kind: &structpb.Value_NumberValue{
								NumberValue: float64(1),
							},
						},
					},
				},
			},
		}

		var factsGathered events.FactsGathered

		err := events.FromEvent(message, &factsGathered)
		suite.Require().NoError(err)
		suite.Equal(expectedFactsGathered.GetAgentId(), factsGathered.GetAgentId())
		suite.Equal(expectedFactsGathered.GetExecutionId(), factsGathered.GetExecutionId())
		suite.Equal(expectedFactsGathered.GetFactsGathered(), factsGathered.GetFactsGathered())

		return nil
	}

	err = suite.RabbitMQAdapter.Listen(handle)
	if err != nil {
		panic(err)
	}

	testsupport.PublishUntilDone(groupCtx, suite.RabbitMQAdapter, "agents", event,
		"failed to publish facts gathering request, will retry")

	err = g.Wait()
	if err != nil {
		panic(err)
	}
}
