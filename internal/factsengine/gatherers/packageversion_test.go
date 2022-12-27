package gatherers_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
)

type PackageVersionTestSuite struct {
	suite.Suite
	mockExecutor *utilsMocks.CommandExecutor
}

func TestPackageVersionTestSuite(t *testing.T) {
	suite.Run(t, new(PackageVersionTestSuite))
}

func (suite *PackageVersionTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
}

func (suite *PackageVersionTestSuite) TestPackageVersionGathererNoArgumentProvided() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "").Return(
		[]byte("rpm: no arguments given for query"), nil)

	p := gatherers.NewPackageVersionGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "package_version",
			CheckID:  "check1",
		},
		{
			Name:     "empty_argument_fact",
			Gatherer: "package_version",
			Argument: "",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "no_argument_fact",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "package-version-missing-argument",
			},
		},
		{
			Name:    "empty_argument_fact",
			CheckID: "check2",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "package-version-missing-argument",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PackageVersionTestSuite) TestPackageVersionGather() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "corosync").Return(
		[]byte("2.4.5"), nil)
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "pacemaker").Return(
		[]byte("2.0.5+20201202.ba59be712"), nil)

	p := gatherers.NewPackageVersionGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync",
			Gatherer: "package_version",
			Argument: "corosync",
			CheckID:  "check1",
		},
		{
			Name:     "pacemaker",
			Gatherer: "package_version",
			Argument: "pacemaker",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "corosync",
			Value:   &entities.FactValueString{Value: "2.4.5"},
			CheckID: "check1",
		},
		{
			Name:    "pacemaker",
			Value:   &entities.FactValueString{Value: "2.0.5+20201202.ba59be712"},
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PackageVersionTestSuite) TestPackageVersionGatherMissingPackageError() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "corosync").Return(
		[]byte("2.4.5"), nil)
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "pacemake").Return(
		[]byte("Error getting version of package: pacemake\n"), errors.New("some error"))

	p := gatherers.NewPackageVersionGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "corosync",
			Gatherer: "package_version",
			Argument: "corosync",
			CheckID:  "check1",
		},
		{
			Name:     "pacemaker",
			Gatherer: "package_version",
			Argument: "pacemake",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "corosync",
			Value:   &entities.FactValueString{Value: "2.4.5"},
			CheckID: "check1",
		},
		{
			Name:  "pacemaker",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error getting version of package: pacemake",
				Type:    "package-version-cmd-error",
			},
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
