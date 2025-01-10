package gatherers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	mocks "github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
	"github.com/trento-project/agent/pkg/factsengine/entities"

	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

type SystemDV2TestSuite struct {
	suite.Suite
}

func TestSystemDV2TestSuite(t *testing.T) {
	suite.Run(t, new(SystemDTestSuite))
}

func (suite *SystemDTestSuite) TestSystemDV2NoArgumentProvided() {
	mockConnector := mocks.NewDbusConnector(suite.T())

	s := gatherers.NewSystemDGathererV2(mockConnector, true)

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

func (suite *SystemDTestSuite) TestSystemDV2Gather() {
	mockConnector := mocks.NewDbusConnector(suite.T())

	corosyncProperties := map[string]interface{}{
		"ActiveState":      "inactive",
		"Description":      "Corosync Cluster Engine",
		"Id":               "corosync.service",
		"LoadState":        "loaded",
		"NeedDaemonReload": "false",
		"UnitFilePreset":   "disabled",
		"UnitFileState":    "disabled",
	}

	pacemakerProperties := map[string]interface{}{
		"ActiveState":      "active",
		"Description":      "Pacemaker High Availability Cluster Manager",
		"Id":               "pacemaker.service",
		"LoadState":        "loaded",
		"NeedDaemonReload": "false",
		"UnitFilePreset":   "enabled",
		"UnitFileState":    "enabled",
	}

	mockConnector.
		On("GetUnitPropertiesContext", mock.Anything, "corosync.service").
		Return(corosyncProperties, nil).
		On("GetUnitPropertiesContext", mock.Anything, "pacemaker.service").
		Return(pacemakerProperties, nil)

	s := gatherers.NewSystemDGathererV2(mockConnector, true)

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync",
			Gatherer: "systemd@v2",
			Argument: "corosync.service",
			CheckID:  "check1",
		},
		{
			Name:     "pacemaker",
			Gatherer: "systemd@v2",
			Argument: "pacemaker.service",
			CheckID:  "check2",
		},
	}

	factResults, err := s.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "corosync",
			Value: &entities.FactValueMap{Value: map[string]entities.FactValue{
				"active_state":       &entities.FactValueString{Value: "inactive"},
				"description":        &entities.FactValueString{Value: "Corosync Cluster Engine"},
				"id":                 &entities.FactValueString{Value: "corosync.service"},
				"load_state":         &entities.FactValueString{Value: "loaded"},
				"need_daemon_reload": &entities.FactValueBool{Value: false},
				"unit_file_preset":   &entities.FactValueString{Value: "disabled"},
				"unit_file_state":    &entities.FactValueString{Value: "disabled"},
			}},
			CheckID: "check1",
		},
		{
			Name: "pacemaker",
			Value: &entities.FactValueMap{Value: map[string]entities.FactValue{
				"active_state":       &entities.FactValueString{Value: "active"},
				"description":        &entities.FactValueString{Value: "Pacemaker High Availability Cluster Manager"},
				"id":                 &entities.FactValueString{Value: "pacemaker.service"},
				"load_state":         &entities.FactValueString{Value: "loaded"},
				"need_daemon_reload": &entities.FactValueBool{Value: false},
				"unit_file_preset":   &entities.FactValueString{Value: "enabled"},
				"unit_file_state":    &entities.FactValueString{Value: "enabled"},
			}},
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SystemDTestSuite) TestSystemDV2GatherNotInitialized() {
	mockConnector := mocks.NewDbusConnector(suite.T())

	s := gatherers.NewSystemDGathererV2(mockConnector, false)

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync",
			Gatherer: "systemd",
			Argument: "corosync.service",
			CheckID:  "check1",
		},
		{
			Name:     "pacemaker",
			Gatherer: "systemd",
			Argument: "pacemaker.service",
			CheckID:  "check2",
		},
	}

	_, err := s.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: systemd-dbus-not-initialized - "+
		"systemd gatherer not initialized properly")
}

func (suite *SystemDTestSuite) TestSystemDV2GatherError() {
	mockConnector := mocks.NewDbusConnector(suite.T())

	mockConnector.On("GetUnitPropertiesContext", mock.Anything, mock.Anything).Return(
		nil, errors.New("error getting properties"))

	s := gatherers.NewSystemDGathererV2(mockConnector, true)

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync",
			Gatherer: "systemd",
			Argument: "corosync",
			CheckID:  "check1",
		},
	}

	factResults, err := s.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "corosync",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "error getting systemd unit properties: argument corosync: error getting properties",
				Type:    "systemd-unit-error",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
