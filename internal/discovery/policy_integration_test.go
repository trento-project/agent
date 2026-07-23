// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package discovery_test

import (
	"context"
	"log/slog"
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
		amqpService = "amqp://guest:guest@localhost:5675" //nolint:gosec
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

func (suite *PolicyIntegrationTestSuite) TestDiscoveryIntegration() {
	agentID := "some-agent"
	// Bounded as a safety net: if the request is never picked up, fail fast
	// with a clear error instead of hanging until the outer test timeout.
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()
	g, groupCtx := errgroup.WithContext(ctx)

	testDiscovery := mocks.NewMockDiscovery(suite.T())
	discoveries := []discovery.Discovery{testDiscovery}

	testDiscovery.
		On("GetID").
		Return("test_discovery")

	testDiscovery.
		On("Discover", mock.Anything).
		Return("discovered", nil).
		Run(func(_ mock.Arguments) {
			ctxCancel()
		}).
		Once()

	g.Go(func() error {
		err := discovery.ListenRequests(groupCtx, agentID, suite.amqpService, discoveries)
		suite.Require().NoError(err)

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

	// The listener goroutine binds its queue asynchronously, so a single
	// publish can race ahead of that binding and be silently dropped.
	// Retry until the request is picked up (or the context above expires)
	// instead of relying on a fixed sleep. Publish errors are logged and
	// retried rather than treated as fatal: transient reconnects on the
	// underlying AMQP connection are expected and should not fail the test.
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for groupCtx.Err() == nil {
		if err := suite.rabbitmqAdapter.Publish("agents", "", event); err != nil {
			slog.Warn("failed to publish discovery request, will retry", "error", err)
		}

		select {
		case <-groupCtx.Done():
		case <-ticker.C:
		}
	}

	err = g.Wait()
	if err != nil {
		panic(err)
	}

	testDiscovery.AssertExpectations(suite.T())
}
