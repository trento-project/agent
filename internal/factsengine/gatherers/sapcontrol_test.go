package gatherers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/factscache"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"

	sapcontrol "github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
	sapControlMocks "github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi/mocks"
)

type SapControlGathererSuite struct {
	suite.Suite
	testFS     afero.Fs
	cache      *factscache.FactsCache
	webService *sapControlMocks.WebServiceConnector
}

func TestSapControlGathererSuite(t *testing.T) {
	suite.Run(t, new(SapControlGathererSuite))
}

func (suite *SapControlGathererSuite) SetupSuite() {
	testFS := afero.NewMemMapFs()
	err := testFS.MkdirAll("/usr/sap/PRD/ASCS00", 0644)
	suite.NoError(err)

	suite.testFS = testFS
}

func (suite *SapControlGathererSuite) SetupTest() {
	suite.cache = factscache.NewFactsCache()
	suite.webService = new(sapControlMocks.WebServiceConnector)
}

func (suite *SapControlGathererSuite) TestSapControlGathererArgumentErrors() {
	gatherer := gatherers.NewSapControlGatherer(suite.webService, suite.testFS, nil)

	fr := []entities.FactRequest{
		{
			Name:     "missing_argument",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "",
		},
		{
			Name:     "unsupported_argument",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "Unsupported",
		},
	}

	expectedFacts := []entities.Fact{
		{
			Name:    "missing_argument",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "missing required argument",
				Type:    "sapcontrol-missing-argument",
			},
		},
		{
			Name:    "unsupported_argument",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "the requested argument is not currently supported: Unsupported",
				Type:    "sapcontrol-unsupported-argument",
			},
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)
}

func (suite *SapControlGathererSuite) TestSapControlGathererEmptyFileSystem() {
	gatherer := gatherers.NewSapControlGatherer(suite.webService, afero.NewMemMapFs(), nil)

	fr := []entities.FactRequest{{
		Name:     "sapcontrol",
		Gatherer: "sapcontrol",
		CheckID:  "check1",
		Argument: "GetProcessList",
	}}

	expectedFacts := []entities.Fact{
		{
			Name:    "sapcontrol",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{},
			},
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)
}

func (suite *SapControlGathererSuite) TestSapControlGathererCacheHit() {
	ctx := context.Background()
	mockWebService := new(sapControlMocks.WebService)
	mockWebService.On("GetProcessList", ctx).Return(&sapcontrol.GetProcessListResponse{
		Processes: []*sapcontrol.OSProcess{
			{
				Name: "process1",
			},
			{
				Name: "process2",
			},
		},
	}, nil)

	suite.webService.On("New", "00").Return(mockWebService).Once()

	gatherer := gatherers.NewSapControlGatherer(suite.webService, suite.testFS, suite.cache)

	fr := []entities.FactRequest{
		{
			Name:     "request1",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "GetProcessList",
		},
		{
			Name:     "request2",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "GetProcessList",
		},
	}

	expectedFacts := []entities.Fact{
		{
			Name:    "request1",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"instance_nr": &entities.FactValueString{Value: "00"},
									"name":        &entities.FactValueString{Value: "ASCS00"},
									"output": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process1"},
												},
											},
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process2"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Error: nil,
		},
		{
			Name:    "request2",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"instance_nr": &entities.FactValueString{Value: "00"},
									"name":        &entities.FactValueString{Value: "ASCS00"},
									"output": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process1"},
												},
											},
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process2"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Error: nil,
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)
	suite.webService.AssertNumberOfCalls(suite.T(), "New", 1)
	mockWebService.AssertNumberOfCalls(suite.T(), "GetProcessList", 1)

	entries := suite.cache.Entries()
	suite.ElementsMatch([]string{"sapcontrol:GetProcessList:PRD:00"}, entries)
}

func (suite *SapControlGathererSuite) TestSapControlGathererMultipleInstaces() {
	ctx := context.Background()
	testFS := afero.NewMemMapFs()
	err := testFS.MkdirAll("/usr/sap/PRD/ASCS00", 0644)
	suite.NoError(err)
	err = testFS.MkdirAll("/usr/sap/PRD/ERS10", 0644)
	suite.NoError(err)
	err = testFS.MkdirAll("/usr/sap/QAS/D01", 0644)
	suite.NoError(err)
	err = testFS.MkdirAll("/usr/sap/QAS/D02", 0644)
	suite.NoError(err)

	mockWebService := new(sapControlMocks.WebService)
	mockWebService.On("GetProcessList", ctx).Return(&sapcontrol.GetProcessListResponse{
		Processes: []*sapcontrol.OSProcess{
			{
				Name: "process1",
			},
			{
				Name: "process2",
			},
		},
	}, nil)

	mockWebServiceError := new(sapControlMocks.WebService)
	mockWebServiceError.On("GetProcessList", ctx).Return(nil, fmt.Errorf("some error"))

	suite.webService.
		On("New", "00").Return(mockWebService).
		On("New", "10").Return(mockWebService).
		On("New", "01").Return(mockWebService).
		On("New", "02").Return(mockWebServiceError)

	gatherer := gatherers.NewSapControlGatherer(suite.webService, testFS, suite.cache)

	fr := []entities.FactRequest{
		{
			Name:     "sapcontrol",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "GetProcessList",
		},
	}

	expectedFacts := []entities.Fact{
		{
			Name:    "sapcontrol",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"instance_nr": &entities.FactValueString{Value: "00"},
									"name":        &entities.FactValueString{Value: "ASCS00"},
									"output": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process1"},
												},
											},
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process2"},
												},
											},
										},
									},
								},
							},
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"instance_nr": &entities.FactValueString{Value: "10"},
									"name":        &entities.FactValueString{Value: "ERS10"},
									"output": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process1"},
												},
											},
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process2"},
												},
											},
										},
									},
								},
							},
						},
					},
					"QAS": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"instance_nr": &entities.FactValueString{Value: "01"},
									"name":        &entities.FactValueString{Value: "D01"},
									"output": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process1"},
												},
											},
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"name": &entities.FactValueString{Value: "process2"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Error: nil,
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)

	entries := suite.cache.Entries()
	expectedEntries := []string{
		"sapcontrol:GetProcessList:PRD:00",
		"sapcontrol:GetProcessList:PRD:10",
		"sapcontrol:GetProcessList:QAS:01",
		"sapcontrol:GetProcessList:QAS:02",
	}
	suite.ElementsMatch(expectedEntries, entries)
}

func (suite *SapControlGathererSuite) TestSapControlGathererGetSystemInstanceList() {
	ctx := context.Background()
	mockWebService := new(sapControlMocks.WebService)
	mockWebService.On("GetSystemInstanceList", ctx).Return(&sapcontrol.GetSystemInstanceListResponse{
		Instances: []*sapcontrol.SAPInstance{
			{
				Hostname: "host1",
			},
			{
				Hostname: "host2",
			},
		},
	}, nil)

	suite.webService.On("New", "00").Return(mockWebService)

	gatherer := gatherers.NewSapControlGatherer(suite.webService, suite.testFS, suite.cache)

	fr := []entities.FactRequest{
		{
			Name:     "sapcontrol",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "GetSystemInstanceList",
		},
	}

	expectedFacts := []entities.Fact{
		{
			Name:    "sapcontrol",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"instance_nr": &entities.FactValueString{Value: "00"},
									"name":        &entities.FactValueString{Value: "ASCS00"},
									"output": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"hostname":    &entities.FactValueString{Value: "host1"},
													"instance_nr": &entities.FactValueInt{Value: 0},
												},
											},
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"hostname":    &entities.FactValueString{Value: "host2"},
													"instance_nr": &entities.FactValueInt{Value: 0},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Error: nil,
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)

	entries := suite.cache.Entries()
	suite.ElementsMatch([]string{"sapcontrol:GetSystemInstanceList:PRD:00"}, entries)
}

func (suite *SapControlGathererSuite) TestSapControlGathererGetVersionInfo() {
	ctx := context.Background()
	mockWebService := new(sapControlMocks.WebService)
	mockWebService.On("GetVersionInfo", ctx).Return(&sapcontrol.GetVersionInfoResponse{
		InstanceVersions: []*sapcontrol.VersionInfo{
			{
				Filename:    "/usr/sap/NWP/ERS10/exe/sapstartsrv",
				VersionInfo: "753, patch 900, changelist 2094654, RKS compatibility level 1, optU (Oct 16 2021, 00:03:15), linuxx86_64",
			},
			{
				Filename:    "/usr/sap/NWP/ERS10/exe/enq_server",
				VersionInfo: "755, patch 905, changelist 2094660, RKS compatibility level 2, optU (Oct 16 2021, 00:03:15), arch",
			},
		},
	}, nil)

	suite.webService.On("New", "00").Return(mockWebService)

	gatherer := gatherers.NewSapControlGatherer(suite.webService, suite.testFS, nil)

	fr := []entities.FactRequest{
		{
			Name:     "sapcontrol",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "GetVersionInfo",
		},
	}

	expectedFacts := []entities.Fact{
		{
			Name:    "sapcontrol",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"instance_nr": &entities.FactValueString{Value: "00"},
									"name":        &entities.FactValueString{Value: "ASCS00"},
									"output": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"filename":                &entities.FactValueString{Value: "/usr/sap/NWP/ERS10/exe/sapstartsrv"},
													"sap_kernel":              &entities.FactValueString{Value: "753"},
													"patch":                   &entities.FactValueString{Value: "900"},
													"changelist":              &entities.FactValueString{Value: "2094654"},
													"rks_compatibility_level": &entities.FactValueString{Value: "1"},
													"build":                   &entities.FactValueString{Value: "optU (Oct 16 2021, 00:03:15)"},
													"architecture":            &entities.FactValueString{Value: "linuxx86_64"},
												},
											},
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"filename":                &entities.FactValueString{Value: "/usr/sap/NWP/ERS10/exe/enq_server"},
													"sap_kernel":              &entities.FactValueString{Value: "755"},
													"patch":                   &entities.FactValueString{Value: "905"},
													"changelist":              &entities.FactValueString{Value: "2094660"},
													"rks_compatibility_level": &entities.FactValueString{Value: "2"},
													"build":                   &entities.FactValueString{Value: "optU (Oct 16 2021, 00:03:15)"},
													"architecture":            &entities.FactValueString{Value: "arch"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Error: nil,
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)
}

func (suite *SapControlGathererSuite) TestSapControlGathererHACheckConfig() {
	ctx := context.Background()
	mockWebService := new(sapControlMocks.WebService)
	mockWebService.On("HACheckConfig", ctx).Return(&sapcontrol.HACheckConfigResponse{
		Checks: []*sapcontrol.HACheck{
			{
				Description: "desc1",
			},
			{
				Description: "desc2",
			},
		},
	}, nil)

	suite.webService.On("New", "00").Return(mockWebService)

	gatherer := gatherers.NewSapControlGatherer(suite.webService, suite.testFS, nil)

	fr := []entities.FactRequest{
		{
			Name:     "sapcontrol",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "HACheckConfig",
		},
	}

	expectedFacts := []entities.Fact{
		{
			Name:    "sapcontrol",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"instance_nr": &entities.FactValueString{Value: "00"},
									"name":        &entities.FactValueString{Value: "ASCS00"},
									"output": &entities.FactValueList{
										Value: []entities.FactValue{
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"description": &entities.FactValueString{Value: "desc1"},
												},
											},
											&entities.FactValueMap{
												Value: map[string]entities.FactValue{
													"description": &entities.FactValueString{Value: "desc2"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Error: nil,
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)
}

func (suite *SapControlGathererSuite) TestSapControlGathererHAGetFailoverConfig() {
	ctx := context.Background()
	mockWebService := new(sapControlMocks.WebService)
	mockWebService.On("HAGetFailoverConfig", ctx).Return(&sapcontrol.HAGetFailoverConfigResponse{
		HAActive: false,
		HANodes:  &[]string{"node1"},
	}, nil)

	suite.webService.On("New", "00").Return(mockWebService)

	gatherer := gatherers.NewSapControlGatherer(suite.webService, suite.testFS, nil)

	fr := []entities.FactRequest{
		{
			Name:     "sapcontrol",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "HAGetFailoverConfig",
		},
	}

	expectedFacts := []entities.Fact{
		{
			Name:    "sapcontrol",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueList{
						Value: []entities.FactValue{
							&entities.FactValueMap{
								Value: map[string]entities.FactValue{
									"instance_nr": &entities.FactValueString{Value: "00"},
									"name":        &entities.FactValueString{Value: "ASCS00"},
									"output": &entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"ha_active":                &entities.FactValueBool{Value: false},
											"ha_product_version":       &entities.FactValueString{Value: ""},
											"ha_sap_interface_version": &entities.FactValueString{Value: ""},
											"ha_documentation":         &entities.FactValueString{Value: ""},
											"ha_active_nodes":          &entities.FactValueString{Value: ""},
											"ha_nodes": &entities.FactValueList{
												Value: []entities.FactValue{
													&entities.FactValueString{Value: "node1"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Error: nil,
		},
	}

	results, err := gatherer.Gather(context.Background(), fr)
	suite.NoError(err)
	suite.EqualValues(expectedFacts, results)
}

func (suite *SapControlGathererSuite) TestSapControlGathererContextCancelled() {

	gatherer := gatherers.NewSapControlGatherer(suite.webService, suite.testFS, nil)

	factsRequest := []entities.FactRequest{
		{
			Name:     "missing_argument",
			Gatherer: "sapcontrol",
			CheckID:  "check1",
			Argument: "",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	factResults, err := gatherer.Gather(ctx, factsRequest)

	suite.Error(err)
	suite.Empty(factResults)
}
