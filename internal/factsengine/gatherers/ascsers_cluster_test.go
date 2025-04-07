package gatherers_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	sapcontrol "github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
	sapControlMocks "github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi/mocks"
	"github.com/trento-project/agent/internal/factsengine/factscache"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type AscsErsClusterTestSuite struct {
	suite.Suite
	cache        *factscache.FactsCache
	mockExecutor *utilsMocks.CommandExecutor
	webService   *sapControlMocks.WebServiceConnector
}

func TestAscsErsClusterTestSuite(t *testing.T) {
	suite.Run(t, new(AscsErsClusterTestSuite))
}

func (suite *AscsErsClusterTestSuite) SetupTest() {
	suite.cache = factscache.NewFactsCache()
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
	suite.webService = new(sapControlMocks.WebServiceConnector)
}

func (suite *AscsErsClusterTestSuite) TestAscsErsClusterGatherCmdNotFound() {
	suite.mockExecutor.On("ExecContext", mock.Anything, "/usr/sbin/cibadmin", "--query", "--local").Return(
		[]byte{}, errors.New("cibadmin not found"))

	p := gatherers.NewAscsErsClusterGatherer(suite.mockExecutor, suite.webService, nil)

	factRequests := []entities.FactRequest{
		{
			Name:     "ascsers",
			Gatherer: "ascsers_cluster",
			Argument: "",
			CheckID:  "check1",
		},
	}

	_, err := p.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: cibadmin-command-error - "+
		"error running cibadmin command: cibadmin not found")
}

func (suite *AscsErsClusterTestSuite) TestAscsErsClusterGatherCacheCastingError() {
	cache := factscache.NewFactsCache()
	_, err := cache.GetOrUpdate("/usr/sbin/cibadmin", func(_ ...interface{}) (interface{}, error) {
		return 1, nil
	})
	suite.NoError(err)

	p := gatherers.NewAscsErsClusterGatherer(suite.mockExecutor, suite.webService, cache)

	factRequests := []entities.FactRequest{
		{
			Name:     "ascsers",
			Gatherer: "ascsers_cluster",
			Argument: "",
			CheckID:  "check1",
		},
	}

	_, err = p.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: ascsers-cluster-decoding-error - "+
		"error decoding cibadmin output: error casting the command output")
}

func (suite *AscsErsClusterTestSuite) TestAscsErsClusterGatherInvalidInstanceName() {
	lFile, _ := os.Open(helpers.GetFixturePath("gatherers/cibadmin_multisid_invalid.xml"))
	content, _ := io.ReadAll(lFile)

	suite.mockExecutor.On("ExecContext", mock.Anything, "/usr/sbin/cibadmin", "--query", "--local").Return(
		content, nil)

	p := gatherers.NewAscsErsClusterGatherer(suite.mockExecutor, suite.webService, nil)

	factRequests := []entities.FactRequest{
		{
			Name:     "ascsers",
			Gatherer: "ascsers_cluster",
			Argument: "",
			CheckID:  "check1",
		},
	}

	_, err := p.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: ascsers-cluster-cib-error - "+
		"error parsing cibadmin output: incorrect InstanceName property value: PRD_ASCS00")
}

func (suite *AscsErsClusterTestSuite) TestAscsErsClusterGatherInvalidInstanceNumber() {
	lFile, _ := os.Open(
		helpers.GetFixturePath("gatherers/cibadmin_multisid_invalid_instance_number.xml"))
	content, _ := io.ReadAll(lFile)

	suite.mockExecutor.On("ExecContext", mock.Anything, "/usr/sbin/cibadmin", "--query", "--local").Return(
		content, nil)

	p := gatherers.NewAscsErsClusterGatherer(suite.mockExecutor, suite.webService, nil)

	factRequests := []entities.FactRequest{
		{
			Name:     "ascsers",
			Gatherer: "ascsers_cluster",
			Argument: "",
			CheckID:  "check1",
		},
	}

	_, err := p.Gather(context.Background(), factRequests)

	suite.EqualError(err, "fact gathering error: ascsers-cluster-cib-error - "+
		"error parsing cibadmin output: "+
		"incorrect instance name within the InstanceName value: 0")
}

func (suite *AscsErsClusterTestSuite) TestAscsErsClusterGather() {
	ctx := context.Background()
	lFile, _ := os.Open(helpers.GetFixturePath("gatherers/cibadmin_multisid.xml"))
	content, _ := io.ReadAll(lFile)

	suite.mockExecutor.On("ExecContext", mock.Anything, "/usr/sbin/cibadmin", "--query", "--local").Return(
		content, nil)

	mockWebServicePRDASCS00 := new(sapControlMocks.WebService)
	mockWebServicePRDASCS00.
		On("GetProcessList", ctx).
		Return(&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Name: "enserver",
				},
			},
		}, nil)

	mockWebServicePRDERS10 := new(sapControlMocks.WebService)
	mockWebServicePRDERS10.
		On("GetProcessList", ctx).
		Return(nil, fmt.Errorf("some error"))

	mockWebServiceDEVASCS01 := new(sapControlMocks.WebService)
	mockWebServiceDEVASCS01.
		On("GetProcessList", ctx).
		Return(nil, fmt.Errorf("some error"))

	mockWebServiceDEVERS10 := new(sapControlMocks.WebService)
	mockWebServiceDEVERS10.
		On("GetProcessList", ctx).
		Return(&sapcontrol.GetProcessListResponse{
			Processes: []*sapcontrol.OSProcess{
				{
					Name: "enq_replicator",
				},
			},
		}, nil)

	suite.webService.
		On("New", "00").
		Return(mockWebServicePRDASCS00).
		Once().
		On("New", "10").
		Return(mockWebServicePRDERS10).
		Once().
		On("New", "01").
		Return(mockWebServiceDEVASCS01).
		Once().
		On("New", "11").
		Return(mockWebServiceDEVERS10).
		Once()

	p := gatherers.NewAscsErsClusterGatherer(suite.mockExecutor, suite.webService, suite.cache)

	factRequests := []entities.FactRequest{
		{
			Name:     "ascsers",
			Gatherer: "ascsers_cluster",
			Argument: "",
			CheckID:  "check1",
		},
	}

	results, err := p.Gather(context.Background(), factRequests)

	// nolint:dupl
	expectedFacts := []entities.Fact{
		{
			Name:    "ascsers",
			CheckID: "check1",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"PRD": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"ensa_version": &entities.FactValueString{Value: "ensa1"},
							"instances": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"resource_group":    &entities.FactValueString{Value: "grp_PRD_ASCS00"},
											"resource_instance": &entities.FactValueString{Value: "rsc_sap_PRD_ASCS00"},
											"name":              &entities.FactValueString{Value: "ASCS00"},
											"instance_number":   &entities.FactValueString{Value: "00"},
											"virtual_hostname":  &entities.FactValueString{Value: "sapascs00"},
											"filesystem_based":  &entities.FactValueBool{Value: true},
											"local":             &entities.FactValueBool{Value: true},
										},
									},
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"resource_group":    &entities.FactValueString{Value: "grp_PRD_ERS10"},
											"resource_instance": &entities.FactValueString{Value: "rsc_sap_PRD_ERS10"},
											"name":              &entities.FactValueString{Value: "ERS10"},
											"instance_number":   &entities.FactValueString{Value: "10"},
											"virtual_hostname":  &entities.FactValueString{Value: "sapers10"},
											"filesystem_based":  &entities.FactValueBool{Value: true},
											"local":             &entities.FactValueBool{Value: false},
										},
									},
								},
							},
						},
					},
					"DEV": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"ensa_version": &entities.FactValueString{Value: "ensa2"},
							"instances": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"resource_group":    &entities.FactValueString{Value: "grp_DEV_ASCS01"},
											"resource_instance": &entities.FactValueString{Value: "rsc_sap_DEV_ASCS01"},
											"name":              &entities.FactValueString{Value: "ASCS01"},
											"instance_number":   &entities.FactValueString{Value: "01"},
											"virtual_hostname":  &entities.FactValueString{Value: "sapascs01"},
											"filesystem_based":  &entities.FactValueBool{Value: false},
											"local":             &entities.FactValueBool{Value: false},
										},
									},
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"resource_group":    &entities.FactValueString{Value: "grp_DEV_ERS11"},
											"resource_instance": &entities.FactValueString{Value: "rsc_sap_DEV_ERS11"},
											"name":              &entities.FactValueString{Value: "ERS11"},
											"instance_number":   &entities.FactValueString{Value: "11"},
											"virtual_hostname":  &entities.FactValueString{Value: "sapers11"},
											"filesystem_based":  &entities.FactValueBool{Value: false},
											"local":             &entities.FactValueBool{Value: true},
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

	suite.NoError(err)
	suite.ElementsMatch(expectedFacts, results)
	suite.webService.AssertNumberOfCalls(suite.T(), "New", 4)

	entries := suite.cache.Entries()
	expectedEntries := []string{
		"/usr/sbin/cibadmin",
		"sapcontrol:GetProcessList:PRD:00",
		"sapcontrol:GetProcessList:PRD:10",
		"sapcontrol:GetProcessList:DEV:01",
		"sapcontrol:GetProcessList:DEV:11",
	}
	suite.ElementsMatch(expectedEntries, entries)
}

func (suite *AscsErsClusterTestSuite) TestAscsErsGathererContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := gatherers.NewAscsErsClusterGatherer(utils.Executor{}, suite.webService, nil)
	factRequests := []entities.FactRequest{
		{
			Name:     "ascsers",
			Gatherer: "ascsers_cluster",
			Argument: "",
			CheckID:  "check1",
		},
	}
	factResults, err := c.Gather(ctx, factRequests)

	suite.Error(err)
	suite.Empty(factResults)
}
