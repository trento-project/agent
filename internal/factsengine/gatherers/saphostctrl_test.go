package gatherers_test // nolint:dupl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
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
		[]byte(" Inst Info : PRD - 00 - myhost-vmhana02 - 753, patch 410, changelist 1908545\n Inst Info : PRD - 01 - anotherhost-vmhana01 - 753, patch 410, changelist 1908545\n"), nil)

	p := gatherers.NewSapHostCtrlGatherer(suite.mockExecutor)

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
			Name:    "ping",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"status":  &entities.FactValueString{Value: "SUCCESS"},
					"elapsed": &entities.FactValueString{Value: "543341"},
				},
			},
		},
		{
			Name:    "listinstances",
			CheckID: "check2",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"changelist": &entities.FactValueString{Value: "1908545"},
					"hostname":   &entities.FactValueString{Value: "myhost-vmhana02"},
					"instance":   &entities.FactValueString{Value: "00"},
					"patch":      &entities.FactValueString{Value: "410"},
					"revision":   &entities.FactValueString{Value: "753"},
					"system":     &entities.FactValueString{Value: "PRD"},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"changelist": &entities.FactValueString{Value: "1908545"},
					"hostname":   &entities.FactValueString{Value: "anotherhost-vmhana01"},
					"instance":   &entities.FactValueString{Value: "01"},
					"patch":      &entities.FactValueString{Value: "410"},
					"revision":   &entities.FactValueString{Value: "753"},
					"system":     &entities.FactValueString{Value: "PRD"},
				}},
			}},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGatherError() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "Ping").Return(
		[]byte("FAILURE\n"), nil)
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "Pong").Return(
		[]byte("some error"), errors.New("some error"))

	p := gatherers.NewSapHostCtrlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "ping",
			Gatherer: "saphostctrl",
			Argument: "Ping",
			CheckID:  "check1",
		},
		{
			Name:     "start_instance",
			Gatherer: "saphostctrl",
			Argument: "StartInstance",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "ping",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error while parsing saphostctrl output: FAILURE\n",
				Type:    "saphostctrl-parse-error",
			},
			CheckID: "check1",
		},
		{
			Name:  "start_instance",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "requested webmethod not whitelisted: StartInstance",
				Type:    "saphostctrl-webmethod-error",
			},
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
