package gatherers_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
)

type SapControlTestSuite struct {
	suite.Suite
	mockExecutor *utilsMocks.CommandExecutor
}

func TestSapControlTestSuite(t *testing.T) {
	suite.Run(t, new(SapControlTestSuite))
}

func (suite *SapControlTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
}

func (suite *SapControlTestSuite) TestSapControlGathererNoArgumentProvided() {
	c := gatherers.NewSapControlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "no_argument_fact",
			Gatherer: "sapcontrol",
		},
		{
			Name:     "empty_argument_fact",
			Gatherer: "sapcontrol",
			Argument: "",
		},
	}

	factResults, err := c.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "no_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "sapcontrol-missing-argument",
			},
		},
		{
			Name:  "empty_argument_fact",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "sapcontrol-missing-argument",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapControlTestSuite) TestSapControlGatherCheckHostAgent() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/sapcontrol", "-nr", "00", "-function", "CheckHostAgent").Return(
		[]byte("SAPHostAgent Installed\n"), nil)

	p := gatherers.NewSapControlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "check_host_agent",
			Gatherer: "sapcontrol",
			Argument: "CheckHostAgent",
			CheckID:  "check1",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "checkhostagent",
			CheckID: "check1",
			Value:   &entities.FactValueBool{Value: true},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapControlTestSuite) TestSapControlGatherGetProcessList() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/sapcontrol", "-nr", "00", "-function", "GetProcessList").Return(
		[]byte(" \n"+
			"09.01.2023 13:14:08\n"+
			"GetProcessList\n"+
			"OK\n"+
			"name, description, dispstatus, textstatus, starttime, elapsedtime, pid\n"+
			"hdbdaemon, HDB Daemon, GREEN, Running, 2022 02 11 18:32:06, 7962:42:02, 7570\n"+
			"hdbcompileserver, HDB Compileserver, GREEN, Running, 2022 02 11 18:32:26, 7962:41:42, 8073\n"+
			"hdbindexserver, HDB Indexserver-PRD, GREEN, Running, 2022 02 11 18:32:28, 7962:41:40, 8356\n"+
			"hdbnameserver, HDB Nameserver, GREEN, Running, 2022 02 11 18:32:07, 7962:42:01, 7588\n"+
			"hdbpreprocessor, HDB Preprocessor, GREEN, Running, 2022 02 11 18:32:26, 7962:41:42, 8076\n"+
			"hdbwebdispatcher, HDB Web Dispatcher, GREEN, Running, 2022 02 11 18:32:44, 7962:41:24, 9240\n"+
			"hdbxsengine, HDB XSEngine-PRD, GREEN, Running, 2022 02 11 18:32:28, 7962:41:40, 8359\n"), nil)

	p := gatherers.NewSapControlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "get_process_list",
			Gatherer: "sapcontrol",
			Argument: "GetProcessList",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "get_process_list",
			CheckID: "check2",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"name":        &entities.FactValueString{Value: "hdbdaemon"},
					"description": &entities.FactValueString{Value: "HDB Daemon"},
					"dispstatus":  &entities.FactValueString{Value: "GREEN"},
					"textstatus":  &entities.FactValueString{Value: "Running"},
					"starttime":   &entities.FactValueString{Value: "2022 02 11 18:32:06"},
					"elapsedtime": &entities.FactValueString{Value: "7962:42:02"},
					"pid":         &entities.FactValueInt{Value: 7570},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"name":        &entities.FactValueString{Value: "hdbcompileserver"},
					"description": &entities.FactValueString{Value: "HDB Compileserver"},
					"dispstatus":  &entities.FactValueString{Value: "GREEN"},
					"textstatus":  &entities.FactValueString{Value: "Running"},
					"starttime":   &entities.FactValueString{Value: "2022 02 11 18:32:26"},
					"elapsedtime": &entities.FactValueString{Value: "7962:41:42"},
					"pid":         &entities.FactValueInt{Value: 8073},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"name":        &entities.FactValueString{Value: "hdbindexserver"},
					"description": &entities.FactValueString{Value: "HDB Indexserver-PRD"},
					"dispstatus":  &entities.FactValueString{Value: "GREEN"},
					"textstatus":  &entities.FactValueString{Value: "Running"},
					"starttime":   &entities.FactValueString{Value: "2022 02 11 18:32:28"},
					"elapsedtime": &entities.FactValueString{Value: "7962:41:40"},
					"pid":         &entities.FactValueInt{Value: 8356},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"name":        &entities.FactValueString{Value: "hdbnameserver"},
					"description": &entities.FactValueString{Value: "HDB Nameserver"},
					"dispstatus":  &entities.FactValueString{Value: "GREEN"},
					"textstatus":  &entities.FactValueString{Value: "Running"},
					"starttime":   &entities.FactValueString{Value: "2022 02 11 18:32:07"},
					"elapsedtime": &entities.FactValueString{Value: "7962:42:01"},
					"pid":         &entities.FactValueInt{Value: 7588},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"name":        &entities.FactValueString{Value: "hdbpreprocessor"},
					"description": &entities.FactValueString{Value: "HDB Preprocessor"},
					"dispstatus":  &entities.FactValueString{Value: "GREEN"},
					"textstatus":  &entities.FactValueString{Value: "Running"},
					"starttime":   &entities.FactValueString{Value: "2022 02 11 18:32:26"},
					"elapsedtime": &entities.FactValueString{Value: "7962:41:42"},
					"pid":         &entities.FactValueInt{Value: 8076},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"name":        &entities.FactValueString{Value: "hdbwebdispatcher"},
					"description": &entities.FactValueString{Value: "HDB Web Dispatcher"},
					"dispstatus":  &entities.FactValueString{Value: "GREEN"},
					"textstatus":  &entities.FactValueString{Value: "Running"},
					"starttime":   &entities.FactValueString{Value: "2022 02 11 18:32:44"},
					"elapsedtime": &entities.FactValueString{Value: "7962:41:24"},
					"pid":         &entities.FactValueInt{Value: 9240},
				}},
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"name":        &entities.FactValueString{Value: "hdbxsengine"},
					"description": &entities.FactValueString{Value: "HDB XSEngine-PRD"},
					"dispstatus":  &entities.FactValueString{Value: "GREEN"},
					"textstatus":  &entities.FactValueString{Value: "Running"},
					"starttime":   &entities.FactValueString{Value: "2022 02 11 18:32:28"},
					"elapsedtime": &entities.FactValueString{Value: "7962:41:40"},
					"pid":         &entities.FactValueInt{Value: 8359},
				}},
			}},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapControlTestSuite) TestSapControlGatherGetSystemInstanceList() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/sapcontrol", "-nr", "00", "-function", "GetSystemInstanceList").Return(
		[]byte(" \n"+
			"09.01.2023 13:14:15\n"+
			"GetSystemInstanceList\n"+
			"OK\n"+
			"hostname, instanceNr, httpPort, httpsPort, startPriority, features, dispstatus\n"+
			"somehost-vmhana02, 0, 50013, 50014, 0.3, HDB|HDB_WORKER, GREEN\n"), nil)

	p := gatherers.NewSapControlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "get_system_instance_list",
			Gatherer: "sapcontrol",
			Argument: "GetSystemInstanceList",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "get_system_instance_list",
			CheckID: "check2",
			Value: &entities.FactValueList{Value: []entities.FactValue{
				&entities.FactValueMap{Value: map[string]entities.FactValue{
					"hostname":      &entities.FactValueString{Value: "somehost-vmhana02"},
					"instanceNr":    &entities.FactValueString{Value: "0"},
					"startPriority": &entities.FactValueString{Value: "0.3"},
					"httpPort":      &entities.FactValueInt{Value: 50013},
					"httpsPort":     &entities.FactValueInt{Value: 50014},
					"features":      &entities.FactValueString{Value: "HDB|HDB_WORKER"},
					"dispstatus":    &entities.FactValueString{Value: "GREEN"},
				}},
			}},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *SapControlTestSuite) TestSapControlGatherError() {
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/sapcontrol", "-nr", "00", "-function", "CheckHostAgent").Return(
		[]byte(""), nil)
	suite.mockExecutor.On("Exec", "/usr/sap/hostctrl/exe/sapcontrol", "-nr", "00", "-function", "GetSystemInstanceList").Return(
		nil, errors.New("some error"))

	p := gatherers.NewSapControlGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "check_host_agent",
			Gatherer: "sapcontrol",
			Argument: "CheckHostAgent",
			CheckID:  "check1",
		},
		{
			Name:     "create_snapshot",
			Gatherer: "sapcontrol",
			Argument: "CreateSnapshot",
			CheckID:  "check2",
		},
		{
			Name:     "get_system_instance_list",
			Gatherer: "sapcontrol",
			Argument: "GetSystemInstanceList",
			CheckID:  "check3",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:  "check_host_agent",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error while parsing sapcontrol output: empty output",
				Type:    "sapcontrol-parse-error",
			},
			CheckID: "check1",
		},
		{
			Name:  "create_snapshot",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "requested webmethod not supported: CreateSnapshot",
				Type:    "sapcontrol-webmethod-error",
			},
			CheckID: "check2",
		},
		{
			Name:  "get_system_instance_list",
			Value: nil,
			Error: &entities.FactGatheringError{
				Message: "error executing sapcontrol command: some error",
				Type:    "sapcontrol-cmd-error",
			},
			CheckID: "check3",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
