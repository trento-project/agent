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
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", "%{VERSION}", "").Return(
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
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", "%{VERSION}", "corosync").Return(
		[]byte("2.4.5"), nil)
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", "%{VERSION}", "pacemaker").Return(
		[]byte("2.0.5+20201202.ba59be712"), nil)
	suite.mockExecutor.On("Exec", "/usr/bin/zypper", "--terse", "versioncmp", "2.4.4", "2.4.5").Return(
		[]byte("-1"), nil)
	suite.mockExecutor.On("Exec", "/usr/bin/zypper", "--terse", "versioncmp", "2.4.5", "2.4.5").Return(
		[]byte("0"), nil)
	suite.mockExecutor.On("Exec", "/usr/bin/zypper", "--terse", "versioncmp", "2.4.6", "2.4.5").Return(
		[]byte("1"), nil)

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
		{
			Name:     "corosync_same_version",
			Gatherer: "package_version",
			Argument: "corosync,2.4.5",
			CheckID:  "check3",
		},
		{
			Name:     "corosync_older_than_installed",
			Gatherer: "package_version",
			Argument: "corosync,2.4.4",
			CheckID:  "check4",
		},
		{
			Name:     "corosync_newer_than_installed",
			Gatherer: "package_version",
			Argument: "corosync,2.4.6",
			CheckID:  "check5",
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
		{
			Name:    "corosync_same_version",
			Value:   &entities.FactValueInt{Value: 0},
			CheckID: "check3",
		},
		{
			Name:    "corosync_older_than_installed",
			Value:   &entities.FactValueInt{Value: -1},
			CheckID: "check4",
		},
		{
			Name:    "corosync_newer_than_installed",
			Value:   &entities.FactValueInt{Value: 1},
			CheckID: "check5",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PackageVersionTestSuite) TestPackageVersionGatherErrors() {
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", "%{VERSION}", "sbd").Return(
		[]byte("package sbd is not installed"), errors.New(""))
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", "%{VERSION}", "pacemake").Return(
		[]byte("package pacemake is not installed"), errors.New(""))
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", "%{VERSION}", "corosync").Return(
		[]byte("2.4.5"), nil)
	suite.mockExecutor.On("Exec", "/usr/bin/zypper", "--terse", "versioncmp", "1.2.3", "2.4.5").Return(
		[]byte(""), errors.New("zypper: command not found"))
	p := gatherers.NewPackageVersionGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "sbd_compare_version",
			Gatherer: "package_version",
			Argument: "sbd,1.2.3",
			CheckID:  "check1",
		},
		{
			Name:     "pacemaker",
			Gatherer: "package_version",
			Argument: "pacemake",
			CheckID:  "check2",
		},
		{
			Name:     "corosync_compare",
			Gatherer: "package_version",
			Argument: "corosync,1.2.3",
			CheckID:  "check3",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "sbd_compare_version",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error while fetching package version: package sbd is not installed",
				Type:    "package-version-rpm-cmd-error",
			},
			CheckID: "check1",
		},
		{
			Name:  "pacemaker",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error while fetching package version: package pacemake is not installed",
				Type:    "package-version-rpm-cmd-error",
			},
			CheckID: "check2",
		},
		{
			Name:  "corosync_compare",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error while executing zypper: zypper: command not found",
				Type:    "package-version-zypper-cmd-error",
			},
			CheckID: "check3",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
