package gatherers

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
	mocks "github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type SBDDumpTestSuite struct {
	suite.Suite
}

func TestSBDDumpTestSuite(t *testing.T) {
	suite.Run(t, new(SBDDumpTestSuite))
}

func (suite *SBDDumpTestSuite) TestSBDDumpGathererMissingFact() {
	mockExecutor := new(mocks.CommandExecutor)

	mockOutputFile, _ := os.Open("../../../test/fixtures/gatherers/sbddump.output")
	mockOutput, _ := ioutil.ReadAll(mockOutputFile)
	mockExecutor.On("Exec", "sbd", "dump", "-d", "/dev/sdj", "dump").Return(mockOutput, nil)

	c := &SBDDumpGatherer{
		executor: mockExecutor,
	}

	factRequests := []FactRequest{
		{
			Name:     "nonexistant_timeout",
			Gatherer: "SBD_dump",
			Argument: "/dev/sdj:Timeout (nonexistant)",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []Fact{}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SBDDumpTestSuite) TestSBDDumpGatherer() {
	mockExecutor := new(mocks.CommandExecutor)

	mockOutputFile, _ := os.Open("../../../test/fixtures/gatherers/sbddump.output")
	mockOutput, _ := ioutil.ReadAll(mockOutputFile)
	mockExecutor.On("Exec", "sbd", "dump", "-d", "/dev/sdj", "dump").Return(mockOutput, nil)

	c := &SBDDumpGatherer{
		executor: mockExecutor,
	}

	factRequests := []FactRequest{
		{
			Name:     "header_version",
			Gatherer: "SBD_dump",
			Argument: "/dev/sdj:Header version",
		},
		{
			Name:     "sector_size",
			Gatherer: "SBD_dump",
			Argument: "/dev/sdj:Sector size",
		},
		{
			Name:     "watchdog_timeout",
			Gatherer: "SBD_dump",
			Argument: "/dev/sdj:Timeout (watchdog)",
		},
		{
			Name:     "allocate_timeout",
			Gatherer: "SBD_dump",
			Argument: "/dev/sdj:Timeout (allocate)",
		},
		{
			Name:     "loop_timeout",
			Gatherer: "SBD_dump",
			Argument: "/dev/sdj:Timeout (loop)",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []Fact{
		{
			Name:  "header_version",
			Value: "2.1",
		},
		{
			Name:  "sector_size",
			Value: "512",
		},
		{
			Name:  "watchdog_timeout",
			Value: "5",
		},
		{
			Name:  "allocate_timeout",
			Value: "2",
		},
		{
			Name:  "loop_timeout",
			Value: "1",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SBDDumpTestSuite) TestSBDDumpCommandNotFound() {
	mockExecutor := new(mocks.CommandExecutor)

	mockExecutor.On("Exec", "sbd", "dump", "-d", "/dev/sdj", "dump").Return(nil, exec.ErrNotFound)

	c := &SBDDumpGatherer{
		executor: mockExecutor,
	}

	factRequests := []FactRequest{
		{
			Name:     "watchdog_timeout",
			Gatherer: "SBD_dump",
			Argument: "/dev/sdj:Timeout (watchdog)",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []Fact{}

	suite.Error(err)
	suite.ElementsMatch(expectedResults, factResults)
}
