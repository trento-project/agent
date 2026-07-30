// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/v3/internal/discovery"
	"github.com/trento-project/agent/v3/internal/discovery/mocks"
	"github.com/trento-project/agent/v3/internal/messaging/testsupport"
	"github.com/trento-project/contracts/go/pkg/events"
	"golang.org/x/sync/errgroup"
)

type PolicyIntegrationTestSuite struct {
	testsupport.RabbitMQIntegrationSuite
}

func TestFactsEngineIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	s := &PolicyIntegrationTestSuite{
		RabbitMQIntegrationSuite: testsupport.RabbitMQIntegrationSuite{
			QueueName:    "test",
			ExchangeName: "trento.discoveries",
			RoutingKey:   "test",
		},
	}

	suite.Run(t, s)
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
		})

	g.Go(func() error {
		err := discovery.ListenRequests(groupCtx, agentID, suite.AMQPService, discoveries)
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

	testsupport.PublishUntilDone(groupCtx, suite.RabbitMQAdapter, "agents", event,
		"failed to publish discovery request, will retry")

	err = g.Wait()
	if err != nil {
		panic(err)
	}

	testDiscovery.AssertExpectations(suite.T())
}
