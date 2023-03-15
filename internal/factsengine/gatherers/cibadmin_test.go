package gatherers_test

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type CibAdminTestSuite struct {
	suite.Suite
	mockExecutor   *utilsMocks.CommandExecutor
	cibAdminOutput []byte
}

func TestCibAdminTestSuite(t *testing.T) {
	suite.Run(t, new(CibAdminTestSuite))
}

func (suite *CibAdminTestSuite) SetupSuite() {
	lFile, _ := os.Open(helpers.GetFixturePath("gatherers/cibadmin.xml"))
	content, _ := io.ReadAll(lFile)

	suite.cibAdminOutput = content
}

func (suite *CibAdminTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
}

func (suite *CibAdminTestSuite) TestCibAdminGatherCmdNotFound() {
	suite.mockExecutor.On("Exec", "cibadmin", "--query", "--local").Return(
		suite.cibAdminOutput, errors.New("cibadmin not found"))

	p := gatherers.NewCibAdminGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "cib",
			Gatherer: "cibadmin",
			Argument: "cib",
			CheckID:  "check1",
		},
	}

	_, err := p.Gather(factRequests)

	suite.EqualError(err, "fact gathering error: cibadmin-command-error - "+
		"error running cibadmin command: cibadmin not found")
}

func (suite *CibAdminTestSuite) TestCibAdminInvalidXML() {
	suite.mockExecutor.On("Exec", "cibadmin", "--query", "--local").Return(
		[]byte("invalid"), nil)

	p := gatherers.NewCibAdminGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "cib",
			Gatherer: "cibadmin",
			Argument: "cib",
			CheckID:  "check1",
		},
	}

	_, err := p.Gather(factRequests)

	suite.EqualError(err, "fact gathering error: cibadmin-decoding-error - "+
		"error decoding cibadmin output: EOF")
}

func (suite *CibAdminTestSuite) TestCibAdminGather() {
	suite.mockExecutor.On("Exec", "cibadmin", "--query", "--local").Return(
		suite.cibAdminOutput, nil)

	p := gatherers.NewCibAdminGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "sid",
			Gatherer: "cibadmin",
			Argument: "cib.configuration.resources.master.0.primitive.0.instance_attributes.nvpair.0.value",
			CheckID:  "check1",
		},
		{
			Name:     "nvpair",
			Gatherer: "cibadmin",
			Argument: "cib.configuration.crm_config.cluster_property_set.0.nvpair.0",
			CheckID:  "check2",
		},
		{
			Name:     "not_found",
			Gatherer: "cibadmin",
			Argument: "cib.not_found.crm_config",
			CheckID:  "check3",
		},
		{
			Name:     "primitives",
			Gatherer: "cibadmin",
			Argument: "cib.configuration.resources.primitive.0",
			CheckID:  "check4",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "sid",
			Value:   &entities.FactValueString{Value: "PRD"},
			CheckID: "check1",
		},
		{
			Name: "nvpair",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"id":    &entities.FactValueString{Value: "cib-bootstrap-options-have-watchdog"},
					"name":  &entities.FactValueString{Value: "have-watchdog"},
					"value": &entities.FactValueBool{Value: true},
				},
			},
			CheckID: "check2",
		},
		{
			Name:    "not_found",
			Value:   nil,
			CheckID: "check3",
			Error: &entities.FactGatheringError{
				Type: "value-not-found",
				Message: "error getting value: requested field value not found: " +
					"cib.not_found.crm_config"},
		},
		{
			Name: "primitives",
			Value: &entities.FactValueMap{
				Value: map[string]entities.FactValue{
					"id":    &entities.FactValueString{Value: "stonith-sbd"},
					"class": &entities.FactValueString{Value: "stonith"},
					"type":  &entities.FactValueString{Value: "external/sbd"},
					"instance_attributes": &entities.FactValueMap{
						Value: map[string]entities.FactValue{
							"id": &entities.FactValueString{Value: "stonith-sbd-instance_attributes"},
							"nvpair": &entities.FactValueList{
								Value: []entities.FactValue{
									&entities.FactValueMap{
										Value: map[string]entities.FactValue{
											"id":    &entities.FactValueString{Value: "stonith-sbd-instance_attributes-pcmk_delay_max"},
											"name":  &entities.FactValueString{Value: "pcmk_delay_max"},
											"value": &entities.FactValueInt{Value: 30},
										},
									},
								},
							},
						},
					},
				},
			},
			CheckID: "check4",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
