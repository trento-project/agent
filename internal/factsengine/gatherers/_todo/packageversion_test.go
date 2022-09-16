package gatherers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/entities"
	mocks "github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type PackageVersionTestSuite struct {
	suite.Suite
	mockExecutor *mocks.CommandExecutor
}

func TestPackageVersionTestSuite(t *testing.T) {
	suite.Run(t, new(PackageVersionTestSuite))
}

func (suite *PackageVersionTestSuite) SetupTest() {
	suite.mockExecutor = new(mocks.CommandExecutor)
}

func (suite *PackageVersionTestSuite) TestPackageVersionGather() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "corosync").Return(
		[]byte("2.4.5"), nil)
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "pacemaker").Return(
		[]byte("2.0.5+20201202.ba59be712"), nil)

	p := NewPackageVersionGatherer(suite.mockExecutor)

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

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "corosync",
			Value:   "2.4.5",
			CheckID: "check1",
		},
		{
			Name:    "pacemaker",
			Value:   "2.0.5+20201202.ba59be712",
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PackageVersionTestSuite) TestPackageVersionGatherError() {
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "corosync").Return(
		[]byte("2.4.5"), nil)
	suite.mockExecutor.On("Exec", "rpm", "-q", "--qf", "%{VERSION}", "pacemake").Return(
		[]byte("package pacemake is not installed\n"), errors.New("some error"))

	p := NewPackageVersionGatherer(suite.mockExecutor)

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

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "corosync",
			Value:   "2.4.5",
			CheckID: "check1",
		},
		{
			Name:    "pacemaker",
			Value:   "package pacemake is not installed\n",
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
