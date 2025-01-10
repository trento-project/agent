package gatherers_test

import (
	"context"
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

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGathererNoArgumentProvided() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "Ping").Return(
		[]byte("SUCCESS (  543341 usec)\n"), nil)

	c := gatherers.NewSapHostCtrlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "saphostctrl",
		},
		{
			Name:     "empty_argument_fact",
			Gatherer: "saphostctrl",
			Argument: "",
		},
	}

	factResults, err := c.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "no_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "saphostctrl-missing-argument",
			},
		},
		{
			Name:  "empty_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "saphostctrl-missing-argument",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGatherListInstances() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "ListInstances").Return(
		[]byte(" Inst Info : S41 - 41 - s41app - 785, patch 50, changelist 2091708\n"+
			" Inst Info : S41 - 40 - s41app - 785, patch 50, changelist 2091708\n"+
			" Inst Info : HS1 - 11 - s41db - 753, patch 819, changelist 2069355\n"), nil)

	p := gatherers.NewSapHostCtrlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "listinstances",
			Gatherer: "saphostctrl",
			Argument: "ListInstances",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "listinstances",
			CheckID: "check2",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"changelist": &entities.FactValueInt{Value: 2091708},
					"hostname":   &entities.FactValueString{Value: "s41app"},
					"instance":   &entities.FactValueString{Value: "41"},
					"patch":      &entities.FactValueInt{Value: 50},
					"sapkernel":  &entities.FactValueInt{Value: 785},
					"system":     &entities.FactValueString{Value: "S41"},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"changelist": &entities.FactValueInt{Value: 2091708},
					"hostname":   &entities.FactValueString{Value: "s41app"},
					"instance":   &entities.FactValueString{Value: "40"},
					"patch":      &entities.FactValueInt{Value: 50},
					"sapkernel":  &entities.FactValueInt{Value: 785},
					"system":     &entities.FactValueString{Value: "S41"},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"changelist": &entities.FactValueInt{Value: 2069355},
					"hostname":   &entities.FactValueString{Value: "s41db"},
					"instance":   &entities.FactValueString{Value: "11"},
					"patch":      &entities.FactValueInt{Value: 819},
					"sapkernel":  &entities.FactValueInt{Value: 753},
					"system":     &entities.FactValueString{Value: "HS1"},
				}},
			}},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGatherPingSuccess() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "Ping").Return(
		[]byte("SUCCESS (  543341 usec)\n"), nil)

	p := gatherers.NewSapHostCtrlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "ping",
			Gatherer: "saphostctrl",
			Argument: "Ping",
			CheckID:  "check1",
		},
	}

	factResults, err := p.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "ping",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"status":  &entities.FactValueString{Value: "SUCCESS"},
					"elapsed": &entities.FactValueInt{Value: 543341},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGatherPingFailed() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "Ping").Return(
		[]byte("FAILED (     497 usec)\n"), nil)

	p := gatherers.NewSapHostCtrlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "ping",
			Gatherer: "saphostctrl",
			Argument: "Ping",
			CheckID:  "check1",
		},
	}

	factResults, err := p.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "ping",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"status":  &entities.FactValueString{Value: "FAILED"},
					"elapsed": &entities.FactValueInt{Value: 497},
				},
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapHostCtrlTestSuite) TestSapHostCtrlGatherError() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "Ping").Return(
		[]byte("Unexpected output\n"), nil)
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/saphostctrl", "-function", "ListInstances").Return(
		nil, errors.New("some error"))

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
		{
			Name:     "list_instances",
			Gatherer: "saphostctrl",
			Argument: "ListInstances",
			CheckID:  "check3",
		},
	}

	factResults, err := p.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "ping",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error while parsing saphostctrl output: Unexpected output\n",
				Type:    "saphostctrl-parse-error",
			},
			CheckID: "check1",
		},
		{
			Name:  "start_instance",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "requested webmethod not supported: StartInstance",
				Type:    "saphostctrl-webmethod-error",
			},
			CheckID: "check2",
		},
		{
			Name:  "list_instances",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error executing saphostctrl command: some error",
				Type:    "saphostctrl-cmd-error",
			},
			CheckID: "check3",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
