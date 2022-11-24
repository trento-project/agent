package gatherers // nolint

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	mocks "github.com/trento-project/agent/pkg/factsengine/gatherers/mocks"
	"github.com/trento-project/agent/test/helpers"
)

type CibAdminTestSuite struct {
	suite.Suite
	mockExecutor   *mocks.CommandExecutor
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
	suite.mockExecutor = new(mocks.CommandExecutor)
}

func (suite *CibAdminTestSuite) TestCibAdminGather() {
	suite.mockExecutor.On("Exec", "cibadmin", "--query", "--local").Return(
		suite.cibAdminOutput, nil)

	p := NewCibAdminGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "instance_number",
			Gatherer: "cibadmin",
			Argument: "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='InstanceNumber']/@value",
			CheckID:  "check1",
		},
		{
			Name:     "sid",
			Gatherer: "cibadmin",
			Argument: "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='SID']/@value",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "instance_number",
			Value:   "00",
			CheckID: "check1",
		},
		{
			Name:    "sid",
			Value:   "PRD",
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *CibAdminTestSuite) TestCibAdminGatherCmdNotFound() {
	suite.mockExecutor.On("Exec", "cibadmin", "--query", "--local").Return(
		suite.cibAdminOutput, errors.New("cibadmin not found"))

	p := NewCibAdminGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "instance_number",
			Gatherer: "cibadmin",
			Argument: "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='InstanceNumber']/@value",
			CheckID:  "check1",
		},
		{
			Name:     "sid",
			Gatherer: "cibadmin",
			Argument: "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='SID']/@value",
			CheckID:  "check2",
		},
	}

	_, err := p.Gather(factRequests)

	suite.EqualError(err, "cibadmin not found")
}

func (suite *CibAdminTestSuite) TestCibAdminGatherError() {
	suite.mockExecutor.On("Exec", "cibadmin", "--query", "--local").Return(
		suite.cibAdminOutput, nil)

	p := NewCibAdminGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "instance_number",
			Gatherer: "cibadmin",
			Argument: "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='InstancNumber']/@value",
			CheckID:  "check1",
		},
		{
			Name:     "sid",
			Gatherer: "cibadmin",
			Argument: "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='SID']/@value",
			CheckID:  "check2",
		},
	}

	factResults, err := p.Gather(factRequests)

	expectedResults := []entities.FactsGatheredItem{
		{
			Name:    "instance_number",
			Value:   "",
			CheckID: "check1",
		},
		{
			Name:    "sid",
			Value:   "PRD",
			CheckID: "check2",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
