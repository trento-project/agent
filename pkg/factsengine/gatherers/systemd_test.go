package gatherers_test

import (
	"errors"
	"testing"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	mocks "github.com/trento-project/agent/pkg/factsengine/gatherers/mocks"

	"github.com/trento-project/agent/pkg/factsengine/gatherers"
)

type SystemDTestSuite struct {
	suite.Suite
}

func TestSystemDTestSuite(t *testing.T) {
	suite.Run(t, new(SystemDTestSuite))
}

func (suite *SystemDTestSuite) TestSystemDGather() {
	mockConnector := new(mocks.DbusConnector)

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

	mockConnector.On("ListUnitsByNamesContext", mock.Anything, []string{"corosync.service", "pacemaker.service"}).Return(
		units, nil)

	s := gatherers.NewSystemDGatherer(mockConnector, true)

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

	factResults, err := s.Gather(factRequests)

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
	mockConnector := new(mocks.DbusConnector)

	s := gatherers.NewSystemDGatherer(mockConnector, false)

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

	_, err := s.Gather(factRequests)

	suite.EqualError(err, "fact gathering error: systemd-dbus-not-initialized - "+
		"systemd gatherer not initialized properly")
}

func (suite *SystemDTestSuite) TestSystemDGatherError() {
	mockConnector := new(mocks.DbusConnector)

	mockConnector.On("ListUnitsByNamesContext", mock.Anything, mock.Anything).Return(
		nil, errors.New("error listing"))

	s := gatherers.NewSystemDGatherer(mockConnector, true)

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

	_, err := s.Gather(factRequests)

	suite.EqualError(err, "fact gathering error: systemd-list-units-error - "+
		"error getting unit states: error listing")
}
