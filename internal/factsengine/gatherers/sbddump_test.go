package gatherers_test

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type SBDDumpTestSuite struct {
	suite.Suite
}

func TestSBDDumpTestSuite(t *testing.T) {
	suite.Run(t, new(SBDDumpTestSuite))
}

func (suite *SBDDumpTestSuite) TestSBDDumpUnableToLoadDevices() {
	mockExecutor := new(utilsMocks.CommandExecutor)
	sbdDumpGatherer := gatherers.NewSBDDumpGatherer(
		mockExecutor,
		helpers.GetFixturePath("discovery/cluster/sbd/sbd_config_invalid"),
	)

	factRequests := []entities.FactRequest{
		{
			Name:     "sbd_devices_dump",
			Gatherer: "sbd_dump",
		},
	}

	gatheredFacts, err := sbdDumpGatherer.Gather(factRequests)

	expectedError := &entities.FactGatheringError{
		Message: "error loading the configured sbd devices: could not parse sbd config file: error on line 1: missing =",
		Type:    "sbd-devices-loading-error",
	}
	suite.EqualError(err, expectedError.Error())
	suite.Empty(gatheredFacts)
}

func (suite *SBDDumpTestSuite) TestSBDDumpUnableToDumpDevice() {
	mockExecutor := new(utilsMocks.CommandExecutor)

	mockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/dev.vdc.sbddump.output"))
	mockOutput, _ := ioutil.ReadAll(mockOutputFile)
	mockExecutor.On("Exec", "sbd", "-d", "/dev/vdb", "dump").Return(nil, errors.New("a failure"))
	mockExecutor.On("Exec", "sbd", "-d", "/dev/vdc", "dump").Return(mockOutput, nil)

	sbdDumpGatherer := gatherers.NewSBDDumpGatherer(
		mockExecutor,
		helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"),
	)

	factRequests := []entities.FactRequest{
		{
			Name:     "sbd_devices_dump",
			Gatherer: "sbd_dump",
		},
		{
			Name:     "another_sbd_devices_dump",
			Gatherer: "sbd_dump",
			Argument: "an-ignored-argument",
		},
	}

	gatheredFacts, err := sbdDumpGatherer.Gather(factRequests)

	expectedFacts := []entities.Fact{
		{
			Name:  "sbd_devices_dump",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error while executing sbd dump: Error while dumping information for device /dev/vdb: a failure",
				Type:    "sbd-dump-command-error",
			},
		},
		{
			Name:  "another_sbd_devices_dump",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error while executing sbd dump: Error while dumping information for device /dev/vdb: a failure",
				Type:    "sbd-dump-command-error",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedFacts, gatheredFacts)
}

func (suite *SBDDumpTestSuite) TestSBDDumpGatherer() {
	mockExecutor := new(utilsMocks.CommandExecutor)

	deviceVDBMockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/dev.vdb.sbddump.output"))
	deviceVDBMockOutput, _ := ioutil.ReadAll(deviceVDBMockOutputFile)

	deviceVDCMockOutputFile, _ := os.Open(helpers.GetFixturePath("gatherers/dev.vdc.sbddump.output"))
	deviceVDCMockOutput, _ := ioutil.ReadAll(deviceVDCMockOutputFile)

	mockExecutor.On("Exec", "sbd", "-d", "/dev/vdb", "dump").Return(deviceVDBMockOutput, nil)
	mockExecutor.On("Exec", "sbd", "-d", "/dev/vdc", "dump").Return(deviceVDCMockOutput, nil)

	sbdDumpGatherer := gatherers.NewSBDDumpGatherer(
		mockExecutor,
		helpers.GetFixturePath("discovery/cluster/sbd/sbd_config"),
	)

	factRequests := []entities.FactRequest{
		{
			Name:     "sbd_devices_dump",
			Gatherer: "sbd_dump",
		},
		{
			Name:     "another_sbd_devices_dump",
			Gatherer: "sbd_dump",
			Argument: "an-ignored-argument",
		},
	}

	factResults, err := sbdDumpGatherer.Gather(factRequests)

	deviceVDBDump := &entities.FactValueMap{Value: map[string]entities.FactValue{
		"header_version":   &entities.FactValueFloat{Value: 2.1},
		"number_of_slots":  &entities.FactValueInt{Value: 188},
		"sector_size":      &entities.FactValueInt{Value: 1024},
		"timeout_allocate": &entities.FactValueInt{Value: 2},
		"timeout_loop":     &entities.FactValueInt{Value: 3},
		"timeout_msgwait":  &entities.FactValueInt{Value: 120},
		"timeout_watchdog": &entities.FactValueInt{Value: 60},
		"uuid":             &entities.FactValueString{Value: "e09c8993-0cba-438d-a4c3-78e91f58ee52"},
	}}

	deviceVDCDump := &entities.FactValueMap{Value: map[string]entities.FactValue{
		"header_version":   &entities.FactValueFloat{Value: 2.1},
		"number_of_slots":  &entities.FactValueInt{Value: 255},
		"sector_size":      &entities.FactValueInt{Value: 512},
		"timeout_allocate": &entities.FactValueInt{Value: 2},
		"timeout_loop":     &entities.FactValueInt{Value: 1},
		"timeout_msgwait":  &entities.FactValueInt{Value: 10},
		"timeout_watchdog": &entities.FactValueInt{Value: 5},
		"uuid":             &entities.FactValueString{Value: "e5b7c05a-1d3c-43d0-827a-9d4dd05ca54a"},
	}}

	expectedResults := []entities.Fact{
		{
			Name: "sbd_devices_dump",
			Value: &entities.FactValueMap{Value: map[string]entities.FactValue{
				"/dev/vdb": deviceVDBDump,
				"/dev/vdc": deviceVDCDump,
			}},
		},
		{
			Name: "another_sbd_devices_dump",
			Value: &entities.FactValueMap{Value: map[string]entities.FactValue{
				"/dev/vdb": deviceVDBDump,
				"/dev/vdc": deviceVDCDump,
			}},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
