//go:build integration_test

package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"

	"github.com/trento-project/agent/internal/factsengine"
	"github.com/trento-project/agent/internal/factsengine/adapters"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/contracts/pkg/events"
)

type FactsEngineIntegrationTestSuite struct {
	suite.Suite
	factsEngineService string
	rabbitmqAdapter    adapters.Adapter
}

func TestFactsEngineIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(FactsEngineIntegrationTestSuite))
}

func (suite *FactsEngineIntegrationTestSuite) SetupSuite() {
	factsEngineService := os.Getenv("RABBITMQ_URL")
	if factsEngineService == "" {
		factsEngineService = "amqp://guest:guest@localhost:5672"
	}

	suite.factsEngineService = factsEngineService
}

func (suite *FactsEngineIntegrationTestSuite) SetupTest() {
	rabbitmqAdapter, err := adapters.NewRabbitMQAdapter(suite.factsEngineService)
	if err != nil {
		panic(err)
	}

	suite.rabbitmqAdapter = rabbitmqAdapter
}

func (suite *FactsEngineIntegrationTestSuite) TearDownTest() {
	if suite.rabbitmqAdapter == nil {
		return
	}

	err := suite.rabbitmqAdapter.Unsubscribe()
	if err != nil {
		panic(err)
	}
}

type FactsEngineIntegrationTestGatherer struct{}

func NewFactsEngineIntegrationTestGatherer() *FactsEngineIntegrationTestGatherer {
	return &FactsEngineIntegrationTestGatherer{}
}

func (s *FactsEngineIntegrationTestGatherer) Gather(requests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	for i, req := range requests {
		fact := entities.Fact{
			Name:    req.Name,
			Value:   fmt.Sprint(i),
			CheckID: req.CheckID,
			Error:   nil,
		}
		facts = append(facts, fact)
	}
	return facts, nil
}

// nolint:nosnakecase
func (suite *FactsEngineIntegrationTestSuite) TestFactsEngineIntegration() {
	agentID := "some-agent"

	engine := factsengine.NewFactsEngine(agentID, suite.factsEngineService)
	engine.AddGatherer("integration", NewFactsEngineIntegrationTestGatherer())

	err := engine.Subscribe()
	if err != nil {
		panic(err)
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
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
	event, err := events.ToEvent(&factGatheringRequested, "", "")
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
					Value: &events.Fact_NumericValue{
						NumericValue: 0,
					},
				},
				{
					CheckId: "check2",
					Name:    "test2",
					Value: &events.Fact_NumericValue{
						NumericValue: 1,
					},
				},
			},
		}
		var factsGathered events.FactsGathered
		err := events.FromEvent(message, &factsGathered)
		suite.NoError(err)
		suite.Equal(expectedFactsGathered.AgentId, factsGathered.AgentId)
		suite.Equal(expectedFactsGathered.ExecutionId, factsGathered.ExecutionId)
		suite.Equal(expectedFactsGathered.FactsGathered, factsGathered.FactsGathered)

		return nil
	}

	err = suite.rabbitmqAdapter.Listen("trento.checks.executions", "trento.checks", "executions", handle)
	if err != nil {
		panic(err)
	}

	err = suite.rabbitmqAdapter.Publish("trento.checks", "agents", "", event)
	if err != nil {
		panic(err)
	}

	err = g.Wait()
	if err != nil {
		panic(err)
	}
}
