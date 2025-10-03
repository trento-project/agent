package discovery_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"log/slog"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/agent/internal/discovery/mocks"
	"github.com/trento-project/contracts/go/pkg/events"
)

type PolicyTestSuite struct {
	suite.Suite
	agentID       string
	discoveries   map[string]discovery.Discovery[interface{}]
	testDiscovery *mocks.MockDiscovery
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}

func (suite *PolicyTestSuite) SetupTest() {
	discoveries := make(map[string]discovery.Discovery[interface{}])
	suite.testDiscovery = mocks.NewMockDiscovery(suite.T())
	discoveries["test_discovery"] = suite.testDiscovery
	suite.agentID = "agent"
	suite.discoveries = discoveries
}

func (suite *PolicyTestSuite) TestPolicyHandleEventWrongMessage() {
	err := discovery.HandleEvent(
		context.Background(),
		[]byte(""),
		suite.agentID,
		suite.discoveries,
	)
	suite.ErrorContains(err, "error getting event type")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventInvalideEvent() {
	event, err := events.ToEvent(&events.FactsGathered{})
	suite.NoError(err)

	err = discovery.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		suite.discoveries,
	)
	suite.EqualError(err, "invalid event type: Trento.Checks.V1.FactsGathered")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventDecodingError() {
	discoveryRequestedEvent := &events.DiscoveryRequested{
		DiscoveryType: "test_discovery",
		Targets:       []string{suite.agentID},
	}
	now := time.Now()
	expiration := now.Add(-60 * time.Minute)
	event, err := events.ToEvent(discoveryRequestedEvent,
		events.WithTime(now),
		events.WithExpiration(expiration))
	suite.NoError(err)

	err = discovery.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		suite.discoveries,
	)
	suite.EqualError(err, "error decoding DiscoveryRequested event: "+
		"cannot decode cloudevent, event expired")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventDiscardAgent() {
	discoveryRequestedEvent := &events.DiscoveryRequested{
		DiscoveryType: "test_discovery",
		Targets:       []string{"agent1", "agent2"},
	}
	event, err := events.ToEvent(discoveryRequestedEvent)
	suite.NoError(err)

	err = discovery.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		suite.discoveries,
	)
	suite.NoError(err)
	slog.Info("print", "test_discovery", suite.discoveries["test_discovery"])
	suite.testDiscovery.AssertNumberOfCalls(suite.T(), "Discover", 0)
}

func (suite *PolicyTestSuite) TestPolicyHandleEventUnknownDiscoveryType() {
	discoveryRequestedEvent := &events.DiscoveryRequested{
		DiscoveryType: "unknown_discovery",
		Targets:       []string{suite.agentID},
	}
	event, err := events.ToEvent(discoveryRequestedEvent)
	suite.NoError(err)

	err = discovery.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		suite.discoveries,
	)
	suite.EqualError(err, "unknown discovery type: unknown_discovery")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventDiscoveryError() {
	discoveryRequestedEvent := &events.DiscoveryRequested{
		DiscoveryType: "test_discovery",
		Targets:       []string{suite.agentID},
	}
	event, err := events.ToEvent(discoveryRequestedEvent)
	suite.NoError(err)

	suite.testDiscovery.
		On("DiscoverAndPublish", mock.Anything).
		Return("", fmt.Errorf("error discovering"))

	err = discovery.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		suite.discoveries,
	)
	suite.EqualError(err, "error during discovery: error discovering")
}

func (suite *PolicyTestSuite) TestPolicyHandleEventDiscovery() {
	discoveryRequestedEvent := &events.DiscoveryRequested{
		DiscoveryType: "test_discovery",
		Targets:       []string{suite.agentID},
	}
	event, err := events.ToEvent(discoveryRequestedEvent)
	suite.NoError(err)

	suite.testDiscovery.
		On("DiscoverAndPublish", mock.Anything).
		Return("discovered", nil)

	err = discovery.HandleEvent(
		context.Background(),
		event,
		suite.agentID,
		suite.discoveries,
	)
	suite.NoError(err)
}
