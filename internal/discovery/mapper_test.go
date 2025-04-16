package discovery_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/discovery"
	"github.com/trento-project/contracts/go/pkg/events"
)

type MapperTestSuite struct {
	suite.Suite
}

func TestMapperTestSuite(t *testing.T) {
	suite.Run(t, new(MapperTestSuite))
}

func (suite *MapperTestSuite) TestDiscoveryRequestedFromEvent() {
	event := events.DiscoveryRequested{
		DiscoveryType: "some_discovery",
		Targets:       []string{"target1", "target2"},
	}

	eventBytes, err := events.ToEvent(&event)
	suite.NoError(err)

	request, err := discovery.DiscoveryRequestedFromEvent(eventBytes)
	expectedRequest := &discovery.DiscoveryRequested{
		DiscoveryType: "some_discovery",
		Targets:       []string{"target1", "target2"},
	}

	suite.NoError(err)
	suite.Equal(expectedRequest, request)
}

func (suite *MapperTestSuite) TestDiscoveryRequestedFromEventError() {
	_, err := discovery.DiscoveryRequestedFromEvent([]byte("error"))
	suite.Error(err)
}
