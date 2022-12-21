package gatherers // nolint:dupl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
)

type SapHostCtrlTestSuite struct {
	suite.Suite
	mockExecutor *utilsMocks.CommandExecutor
}

func TestSapHostCtrlTestSuite(t *testing.T) {
	suite.Run(t, new(SapHostCtrlTestSuite))
}

func (suite *SapHostCtrlTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
}

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGather() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "Ping").Return(
		[]byte("SUCCESS (  543341 usec)\n"), nil)
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "ListInstances").Return(
		[]byte(" Inst Info : PRD - 00 - myhost-vmhana02 - 753, patch 410, changelist 1908545\n"), nil)

	p := NewSapHostCtrlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "ping",
			Gatherer: "saphostctrl",
			Argument: "Ping",
			CheckID:  "check1",
		},
		{
			Name:     "listinstances",
			Gatherer: "saphostctrl",
			Argument: "ListInstances",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name: "ping",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"status":  &entities.FactValueString{Value: "SUCCESS"},
					"elapsed": &entities.FactValueString{Value: "543341"},
				},
			},
			CheckID: "check1",
		},
		{
			Name: "listinstances",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"system":     &entities.FactValueString{Value: "PRD"},
					"instance":   &entities.FactValueString{Value: "00"},
					"hostname":   &entities.FactValueString{Value: "myhost-vmhana02"},
					"revision":   &entities.FactValueString{Value: "753"},
					"patch":      &entities.FactValueString{Value: "410"},
					"changelist": &entities.FactValueString{Value: "1908545"},
				},
			},
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGatherError() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "Ping").Return(
		[]byte("SUCCESS (  543341 usec)\n"), nil)
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "Pong").Return(
		[]byte("some error"), errors.New("some error"))

	p := NewSapHostCtrlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
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

	expectedResults := []entities.Fact{
		{
			Name: "ping",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"status":  &entities.FactValueString{Value: "SUCCESS"},
					"elapsed": &entities.FactValueString{Value: "543341"},
				},
			},
			CheckID: "check1",
		},
		{
			Name:  "pong",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "unsupported saphostctrl function: Pong",
				Type:    "saphostctrl-func-error",
			},
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
