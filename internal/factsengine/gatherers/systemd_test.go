package gatherers

import (
	"errors"
	"testing"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
	mocks "github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
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

	s := &SystemDGatherer{
		dbusConnnector: mockConnector,
		initialized:    true,
	}

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

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "corosync",
			Value:   "active",
			CheckID: "check1",
		},
		{
			Name:    "pacemaker",
			Value:   "inactive",
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SystemDTestSuite) TestSystemDGatherNotInitialized() {
	mockConnector := new(mocks.DbusConnector)

	s := &SystemDGatherer{
		dbusConnnector: mockConnector,
		initialized:    false,
	}

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

	gatheredFacts, err := s.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "corosync",
			Value:   nil,
			CheckID: "check1",
			Error: &entities.FactGatheringError{
				Message: "systemd gatherer not initialized properly",
				Type:    "systemd-gatherer-not-initialized",
			},
		},
		{
			Name:    "pacemaker",
			Value:   nil,
			CheckID: "check2",
			Error: &entities.FactGatheringError{
				Message: "systemd gatherer not initialized properly",
				Type:    "systemd-gatherer-not-initialized",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, gatheredFacts)
}

func (suite *SystemDTestSuite) TestSystemDGatherError() {
	mockConnector := new(mocks.DbusConnector)

	mockConnector.On("ListUnitsByNamesContext", mock.Anything, mock.Anything).Return(
		nil, errors.New("error listing"))

	s := &SystemDGatherer{
		dbusConnnector: mockConnector,
		initialized:    true,
	}

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

	gatheredFacts, err := s.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "corosync",
			Value:   nil,
			CheckID: "check1",
			Error: &entities.FactGatheringError{
				Message: "error getting unit states: error listing",
				Type:    "systemd-list-units-error",
			},
		},
		{
			Name:    "pacemaker",
			Value:   nil,
			CheckID: "check2",
			Error: &entities.FactGatheringError{
				Message: "error getting unit states: error listing",
				Type:    "systemd-list-units-error",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, gatheredFacts)
}
