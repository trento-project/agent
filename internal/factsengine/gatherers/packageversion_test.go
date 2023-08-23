package gatherers_test

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

const packageVersionQueryFormat = "VERSION=%{VERSION}\nINSTALLTIME=%{INSTALLTIME}\n---\n"

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
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, "").Return(
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
	corosyncMockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/rpm-query.corosync.output"))
	corosyncVersionMockOutput, _ := io.ReadAll(corosyncMockOutputFile)
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, "corosync").
		Return(corosyncVersionMockOutput, nil)

	pacemakerMockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/rpm-query.pacemaker.output"))
	pacemakerVersionMockOutput, _ := io.ReadAll(pacemakerMockOutputFile)
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, "pacemaker").
		Return(pacemakerVersionMockOutput, nil)

	multiversionsMockOutputFile, _ :=
		os.Open(helpers.GetFixturePath("gatherers/rpm-query-multi-versions.variant-1.output"))
	multiversionsVersionMockOutput, _ := io.ReadAll(multiversionsMockOutputFile)
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, "sbd").
		Return(multiversionsVersionMockOutput, nil)

	multiversionsVariantMockOutputFile, _ :=
		os.Open(helpers.GetFixturePath("gatherers/rpm-query-multi-versions.variant-2.output"))
	multiversionsVariantVersionMockOutput, _ := io.ReadAll(multiversionsVariantMockOutputFile)
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, "awk").
		Return(multiversionsVariantVersionMockOutput, nil)

	suite.mockExecutor.On("Exec", "/usr/bin/zypper", "--terse", "versioncmp", "2.4.4", "2.4.5").Return(
		[]byte("-1\n"), nil)
	suite.mockExecutor.On("Exec", "/usr/bin/zypper", "--terse", "versioncmp", "2.4.5", "2.4.5").Return(
		[]byte("0\n"), nil)
	suite.mockExecutor.On("Exec", "/usr/bin/zypper", "--terse", "versioncmp", "2.4.6", "2.4.5").Return(
		[]byte("1\n"), nil)

	versionComparisonOutputWithWarningFile, _ :=
		os.Open(helpers.GetFixturePath("gatherers/versioncmp-with-warning.output"))
	versionComparisonOutputWithWarning, _ := io.ReadAll(versionComparisonOutputWithWarningFile)
	suite.mockExecutor.On("Exec", "/usr/bin/zypper", "--terse", "versioncmp", "1.5.2", "1.5.2").Return(
		versionComparisonOutputWithWarning, nil)

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
		{
			Name:     "sbd_multiversions",
			Gatherer: "package_version",
			Argument: "sbd",
			CheckID:  "check6",
		},
		{
			Name:     "awk_multiversions",
			Gatherer: "package_version",
			Argument: "awk",
			CheckID:  "check7",
		},
		{
			Name:     "sbd_same_version_with_warning",
			Gatherer: "package_version",
			Argument: "sbd,1.5.2",
			CheckID:  "check8",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "corosync",
			Value: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"version": &entities.FactValueString{Value: "2.4.5"},
						},
					},
				},
			},
			CheckID: "check1",
		},
		{
			Name: "pacemaker",
			Value: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"version": &entities.FactValueString{Value: "2.0.5+20201202.ba59be712"},
						},
					},
				},
			},
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
		{
			Name: "sbd_multiversions",
			Value: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"version": &entities.FactValueString{Value: "1.5.2"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"version": &entities.FactValueString{Value: "1.5.1"},
						},
					},
				},
			},
			CheckID: "check6",
		},
		{
			Name: "awk_multiversions",
			Value: &entities.FactValueList{
				Value: []entities.FactValue{
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"version": &entities.FactValueString{Value: "1.5.1"},
						},
					},
					&entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"version": &entities.FactValueString{Value: "1.5.2"},
						},
					},
				},
			},
			CheckID: "check7",
		},
		{
			Name:    "sbd_same_version_with_warning",
			Value:   &entities.FactValueInt{Value: 0},
			CheckID: "check8",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PackageVersionTestSuite) TestPackageVersionGatherErrors() {
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, "sbd").
		Return([]byte("package sbd is not installed"), errors.New(""))
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, "pacemake").
		Return([]byte("package pacemake is not installed"), errors.New(""))

	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/rpm-query.corosync.output"))
	mockOutput, _ := io.ReadAll(mockOutputFile)
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, "corosync").
		Return(mockOutput, nil)

	invalidDateMockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/rpm-query-invalid-date.output"))
	invalidDateMockOutput, _ := io.ReadAll(invalidDateMockOutputFile)
	suite.mockExecutor.On("Exec", "/usr/bin/rpm", "-q", "--qf", packageVersionQueryFormat, "another_package").
		Return(invalidDateMockOutput, nil)

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
		{
			Name:     "invalid_date",
			Gatherer: "package_version",
			Argument: "another_package",
			CheckID:  "check4",
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
		{
			Name:  "invalid_date",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error while fetching package version: Unable to parse package installation timestamp to an integer: " +
					"strconv.ParseInt: parsing \"an invalid date\": invalid syntax",
				Type: "package-version-rpm-cmd-error",
			},
			CheckID: "check4",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
