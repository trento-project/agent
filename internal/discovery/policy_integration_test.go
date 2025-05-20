package discovery_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/mocks"
	"github.com/trento-project/agent/internal/messaging"
	"github.com/trento-project/contracts/go/pkg/events"
	"golang.org/x/sync/errgroup"
)

type PolicyIntegrationTestSuite struct {
	suite.Suite
	amqpService     string
	rabbitmqAdapter messaging.Adapter
}

func TestFactsEngineIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(PolicyIntegrationTestSuite))
}

func (suite *PolicyIntegrationTestSuite) SetupSuite() {
	amqpService := os.Getenv("RABBITMQ_URL")
	if amqpService == "" {
		amqpService = "amqp://guest:guest@localhost:5675"
	}

	suite.amqpService = amqpService
}

func (suite *PolicyIntegrationTestSuite) SetupTest() {
	rabbitmqAdapter, err := messaging.NewRabbitMQAdapter(
		suite.amqpService,
		"test",
		"trento.discoveries",
		"test",
	)
	if err != nil {
		panic(err)
	}

	suite.rabbitmqAdapter = rabbitmqAdapter
}

func (suite *PolicyIntegrationTestSuite) TearDownTest() {
	if suite.rabbitmqAdapter == nil {
		return
	}

	err := suite.rabbitmqAdapter.Unsubscribe()
	if err != nil {
		panic(err)
	}
}

// nolint:nosnakecase
func (suite *PolicyIntegrationTestSuite) TestDiscoveryIntegration() {
	agentID := "some-agent"
	ctx, ctxCancel := context.WithCancel(context.Background())
	g, groupCtx := errgroup.WithContext(ctx)

	testDiscovery := mocks.NewDiscovery(suite.T())
	discoveries := []discovery.Discovery{testDiscovery}

	testDiscovery.
		On("GetID").
		Return("test_discovery")

	testDiscovery.
		On("Discover", mock.Anything).
		Return("discovered", nil).
		Run(func(_ mock.Arguments) {
			ctxCancel()
		})

	g.Go(func() error {
		err := discovery.ListenRequests(groupCtx, agentID, suite.amqpService, discoveries)
		suite.NoError(err)
		return err
	})

	discoveryRequested := events.DiscoveryRequested{
		DiscoveryType: "test_discovery",
		Targets:       []string{"some-agent"},
	}
	event, err := events.ToEvent(&discoveryRequested, events.WithSource(""))
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

	testDiscovery.AssertExpectations(suite.T())
}
