package gatherers // nolint:dupl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	mocks "github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type SapHostCtrlTestSuite struct {
	suite.Suite
	mockExecutor *mocks.CommandExecutor
}

func TestSapHostCtrlTestSuite(t *testing.T) {
	suite.Run(t, new(SapHostCtrlTestSuite))
}

func (suite *SapHostCtrlTestSuite) SetupTest() {
	suite.mockExecutor = new(mocks.CommandExecutor)
}

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGather() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-nr", "00", "-function", "Ping").Return(
		[]byte("SUCCESS (  543341 usec)\n"), nil)
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-nr", "00", "-function", "Pong").Return(
		[]byte("ERROR"), nil)

	p := NewSapHostCtrlGatherer(suite.mockExecutor)

	factRequests := []FactRequest{
		{
			Name:     "ping",
			Gatherer: "saphostctrl",
			Argument: "Ping",
			CheckID:  "check1",
		},
		{
			Name:     "pong",
			Gatherer: "saphostctrl",
			Argument: "Pong",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []Fact{
		{
			Name:    "ping",
			Value:   "SUCCESS (  543341 usec)\n",
			CheckID: "check1",
		},
		{
			Name:    "pong",
			Value:   "ERROR",
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGatherError() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-nr", "00", "-function", "Ping").Return(
		[]byte("SUCCESS (  543341 usec)\n"), nil)
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-nr", "00", "-function", "Pong").Return(
		[]byte("some error"), errors.New("some error"))

	p := NewSapHostCtrlGatherer(suite.mockExecutor)

	factRequests := []FactRequest{
		{
			Name:     "ping",
			Gatherer: "saphostctrl",
			Argument: "Ping",
			CheckID:  "check1",
		},
		{
			Name:     "pong",
			Gatherer: "saphostctrl",
			Argument: "Pong",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []Fact{
		{
			Name:    "ping",
			Value:   "SUCCESS (  543341 usec)\n",
			CheckID: "check1",
		},
		{
			Name:    "pong",
			Value:   "some error",
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
