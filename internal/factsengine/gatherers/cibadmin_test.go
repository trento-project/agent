package gatherers // nolint

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	mocks "github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type CibAdminTestSuite struct {
	suite.Suite
	cibAdminOutput []byte
}

func TestCibAdminTestSuite(t *testing.T) {
	suite.Run(t, new(CibAdminTestSuite))
}

func (suite *CibAdminTestSuite) SetupSuite() {
	lFile, _ := os.Open("../../../test/fixtures/gatherers/cibadmin.xml")
	content, _ := io.ReadAll(lFile)

	suite.cibAdminOutput = content
}

func (suite *CibAdminTestSuite) TestCibAdminGather() {
	mockExecutor := new(mocks.CommandExecutor)

	mockExecutor.On("Exec", "cibadmin", "--query", "--local").Return(
		suite.cibAdminOutput, nil)

	p := &CibAdminGatherer{
		executor: mockExecutor,
	}

	factRequests := []FactRequest{
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

	expectedResults := []Fact{
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
	mockExecutor := new(mocks.CommandExecutor)

	mockExecutor.On("Exec", "cibadmin", "--query", "--local").Return(
		suite.cibAdminOutput, errors.New("cibadmin not found"))

	p := &CibAdminGatherer{
		executor: mockExecutor,
	}

	factRequests := []FactRequest{
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
	mockExecutor := new(mocks.CommandExecutor)

	mockExecutor.On("Exec", "cibadmin", "--query", "--local").Return(
		suite.cibAdminOutput, nil)

	p := &CibAdminGatherer{
		executor: mockExecutor,
	}

	factRequests := []FactRequest{
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

	expectedResults := []Fact{
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
