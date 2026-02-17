package gatherers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/factsengine/entities"

	"github.com/trento-project/agent/internal/core/dbus/mocks"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

type SystemDTestSuite struct {
	suite.Suite
	mockConnector *mocks.MockConnector
}

func TestSystemDTestSuite(t *testing.T) {
	suite.Run(t, new(SystemDTestSuite))
}
func (suite *SystemDTestSuite) SetupTest() {
	suite.mockConnector = mocks.NewMockConnector(suite.T())
}

func (suite *SystemDTestSuite) TestSystemDNoArgumentProvided() {
	suite.mockConnector.On("ListUnitsByNamesContext", mock.Anything, []string{}).Return(
		nil, nil)

	s := gatherers.NewSystemDGatherer(suite.mockConnector, true)

	factRequests := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "systemd",
			CheckID:  "check1",
		},
		{
			Name:     "empty_argument_fact",
			Gatherer: "systemd",
			Argument: "",
			CheckID:  "check2",
		},
	}

	factResults, err := s.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "no_argument_fact",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "systemd-missing-argument",
			},
		},
		{
			Name:    "empty_argument_fact",
			CheckID: "check2",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "systemd-missing-argument",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SystemDTestSuite) TestSystemDGather() {
	units := []dbus.UnitStatus{
		{
			Name:        "corosync.service",
			ActiveState: "active",
		},
		{
			Name:        "pacemaker.service",
			ActiveState: "inactive",
		},
	}

	suite.mockConnector.On("ListUnitsByNamesContext", mock.Anything, []string{"corosync.service", "pacemaker.service"}).Return(
		units, nil)

	s := gatherers.NewSystemDGatherer(suite.mockConnector, true)

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync",
			Gatherer: "systemd",
			Argument: "corosync",
			CheckID:  "check1",
		},
		{
			Name:     "pacemaker",
			Gatherer: "systemd",
			Argument: "pacemaker",
			CheckID:  "check2",
		},
	}

	factResults, err := s.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "corosync",
			Value:   &entities.FactValueString{Value: "active"},
			CheckID: "check1",
		},
		{
			Name:    "pacemaker",
			Value:   &entities.FactValueString{Value: "inactive"},
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SystemDTestSuite) TestSystemDGatherNotInitialized() {
	s := gatherers.NewSystemDGatherer(suite.mockConnector, false)

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync",
			Gatherer: "systemd",
			Argument: "corosync",
			CheckID:  "check1",
		},
		{
			Name:     "pacemaker",
			Gatherer: "systemd",
			Argument: "pacemaker",
			CheckID:  "check2",
		},
	}

	_, err := s.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: systemd-dbus-not-initialized - "+
		"systemd gatherer not initialized properly")
}

func (suite *SystemDTestSuite) TestSystemDGatherError() {
	suite.mockConnector.On("ListUnitsByNamesContext", mock.Anything, mock.Anything).Return(
		nil, errors.New("error listing"))

	s := gatherers.NewSystemDGatherer(suite.mockConnector, true)

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync",
			Gatherer: "systemd",
			Argument: "corosync",
			CheckID:  "check1",
		},
		{
			Name:     "pacemaker",
			Gatherer: "systemd",
			Argument: "pacemaker",
			CheckID:  "check2",
		},
	}

	_, err := s.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: systemd-list-units-error - "+
		"error getting unit states: error listing")
}

func (suite *SystemDTestSuite) TestSystemDContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())

	connector, err := dbus.NewWithContext(ctx)
	if err != nil {
		suite.T().Fatal(err)
	}

	cancel()
	gatherer := gatherers.NewSystemDGatherer(connector, true)

	factsRequest := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "systemd",
			CheckID:  "check1",
		}}

	factResults, err := gatherer.Gather(ctx, factsRequest)

	suite.Error(err)
	suite.Empty(factResults)
}

func (suite *SystemDTestSuite) TestSystemDContextCancelledLongRunning() {
	ctx, cancel := context.WithCancel(context.Background())

	suite.mockConnector.On("ListUnitsByNamesContext", mock.Anything, []string{}).Return(
		nil, nil).Run(func(_ mock.Arguments) {
		time.Sleep(1 * time.Second)
	}).Once()

	gatherer := gatherers.NewSystemDGatherer(suite.mockConnector, true)

	factsRequest := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "systemd",
			CheckID:  "check1",
		}}

	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	factResults, err := gatherer.Gather(ctx, factsRequest)

	suite.Error(err)
	suite.Empty(factResults)
}
