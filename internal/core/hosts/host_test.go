package hosts_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/core/hosts"
	"github.com/trento-project/agent/internal/core/hosts/systemd/mocks"
)

type HostTestSuite struct {
	suite.Suite
	dbusMock *mocks.MockDbusConnector
}

func TestHost(t *testing.T) {
	suite.Run(t, new(HostTestSuite))
}

func (suite *HostTestSuite) SetupTest() {
	suite.dbusMock = mocks.NewMockDbusConnector(suite.T())
}

func (suite *HostTestSuite) TestDefaultUnitInfoOnDbusConnectionError() {
	ctx := context.Background()

	// nil dbus mock simulates an error in creating the dbus connection
	result := hosts.GetSystemdUnitsStatusWithCustomDbus(ctx, nil, []string{"pacemaker.service"})

	expectedSystemdUnits := []hosts.UnitInfo{
		{
			Name:          "pacemaker.service",
			UnitFileState: "unknown",
		},
	}

	suite.EqualValues(expectedSystemdUnits, result)
}

func (suite *HostTestSuite) TestUnableToGetProperties() {
	ctx := context.Background()

	getPropertiesCall := suite.dbusMock.
		On("GetUnitPropertiesContext", ctx, "pacemaker.service").
		Return(nil, fmt.Errorf("error getting properties"))
	suite.dbusMock.On("Close").
		Return().
		Once().
		NotBefore(getPropertiesCall)

	result := hosts.GetSystemdUnitsStatusWithCustomDbus(ctx, suite.dbusMock, []string{"pacemaker.service"})

	expectedSystemdUnits := []hosts.UnitInfo{
		{
			Name:          "pacemaker.service",
			UnitFileState: "unknown",
		},
	}

	suite.EqualValues(expectedSystemdUnits, result)
}

func (suite *HostTestSuite) TestAbleToGetPartialUnitsInfo() {
	ctx := context.Background()

	units := []string{"pacemaker.service", "another.service", "yet.another.service"}

	getPacemakerPropertiesCall := suite.dbusMock.
		On("GetUnitPropertiesContext", ctx, "pacemaker.service").
		Return(nil, fmt.Errorf("error getting properties"))
	getAnotherServicePropertiesCall := suite.dbusMock.
		On("GetUnitPropertiesContext", ctx, "another.service").
		Return(map[string]any{"UnitFileState": "enabled"}, nil).
		NotBefore(getPacemakerPropertiesCall)
	getYetAnotherServicePropertiesCall := suite.dbusMock.
		On("GetUnitPropertiesContext", ctx, "yet.another.service").
		Return(map[string]any{"UnitFileState": "disabled"}, nil).
		NotBefore(getAnotherServicePropertiesCall)
	suite.dbusMock.On("Close").
		Return().
		Once().
		NotBefore(getYetAnotherServicePropertiesCall)

	result := hosts.GetSystemdUnitsStatusWithCustomDbus(ctx, suite.dbusMock, units)

	expectedSystemdUnits := []hosts.UnitInfo{
		{
			Name:          "pacemaker.service",
			UnitFileState: "unknown",
		},
		{
			Name:          "another.service",
			UnitFileState: "enabled",
		},
		{
			Name:          "yet.another.service",
			UnitFileState: "disabled",
		},
	}
	suite.EqualValues(expectedSystemdUnits, result)
}
