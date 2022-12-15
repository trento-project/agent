package gatherers

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
)

type SBDDumpTestSuite struct {
	suite.Suite
}

func TestSBDDumpTestSuite(t *testing.T) {
	suite.Run(t, new(SBDDumpTestSuite))
}

func (suite *SBDDumpTestSuite) TestSBDDumpGathererMissingFact() {
	mockExecutor := new(utilsMocks.CommandExecutor)

	mockOutputFile, _ := os.Open("../../../test/fixtures/gatherers/sbddump.output")
	mockOutput, _ := ioutil.ReadAll(mockOutputFile)
	mockExecutor.On("Exec", "sbd", "dump", "-d", "/dev/sdj", "dump").Return(mockOutput, nil)

	c := &SBDDumpGatherer{
		executor: mockExecutor,
	}

	factRequests := []entities.FactRequest{
		{
			Name:     "nonexistant_timeout",
			Gatherer: "sbd_dump",
			Argument: "/dev/sdj:Timeout (nonexistant)",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SBDDumpTestSuite) TestSBDDumpGatherer() {
	mockExecutor := new(utilsMocks.CommandExecutor)

	mockOutputFile, _ := os.Open("../../../test/fixtures/gatherers/sbddump.output")
	mockOutput, _ := ioutil.ReadAll(mockOutputFile)
	mockExecutor.On("Exec", "sbd", "dump", "-d", "/dev/sdj", "dump").Return(mockOutput, nil)

	c := &SBDDumpGatherer{
		executor: mockExecutor,
	}

	factRequests := []entities.FactRequest{
		{
			Name:     "header_version",
			Gatherer: "sbd_dump",
			Argument: "/dev/sdj:Header version",
		},
		{
			Name:     "sector_size",
			Gatherer: "sbd_dump",
			Argument: "/dev/sdj:Sector size",
		},
		{
			Name:     "watchdog_timeout",
			Gatherer: "sbd_dump",
			Argument: "/dev/sdj:Timeout (watchdog)",
		},
		{
			Name:     "allocate_timeout",
			Gatherer: "sbd_dump",
			Argument: "/dev/sdj:Timeout (allocate)",
		},
		{
			Name:     "loop_timeout",
			Gatherer: "sbd_dump",
			Argument: "/dev/sdj:Timeout (loop)",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "header_version",
			Value: &entities.FactValueString{Value: "2.1"},
		},
		{
			Name:  "sector_size",
			Value: &entities.FactValueString{Value: "512"},
		},
		{
			Name:  "watchdog_timeout",
			Value: &entities.FactValueString{Value: "5"},
		},
		{
			Name:  "allocate_timeout",
			Value: &entities.FactValueString{Value: "2"},
		},
		{
			Name:  "loop_timeout",
			Value: &entities.FactValueString{Value: "1"},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SBDDumpTestSuite) TestSBDDumpCommandNotFound() {
	mockExecutor := new(utilsMocks.CommandExecutor)

	mockExecutor.On("Exec", "sbd", "dump", "-d", "/dev/sdj", "dump").Return(nil, exec.ErrNotFound)

	c := &SBDDumpGatherer{
		executor: mockExecutor,
	}

	factRequests := []entities.FactRequest{
		{
			Name:     "watchdog_timeout",
			Gatherer: "sbd_dump",
			Argument: "/dev/sdj:Timeout (watchdog)",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{}

	suite.Error(err)
	suite.ElementsMatch(expectedResults, factResults)
}
