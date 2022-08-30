package gatherers

import (
	"testing"

	"github.com/stretchr/testify/suite"
	mocks "github.com/trento-project/agent/internal/utils/mocks"
)

type PasswordTestSuite struct {
	suite.Suite
	mockExecutor *mocks.CommandExecutor
}

func TestPasswordTestSuite(t *testing.T) {
	suite.Run(t, new(PasswordTestSuite))
}

func (suite *PasswordTestSuite) SetupTest() {
	suite.mockExecutor = new(mocks.CommandExecutor)
}

func (suite *PasswordTestSuite) TestPasswordGatherEqual() {
	shadow := []byte("hacluster:$6$WFEgPAefduOyvLCN$MprO90En7b/" +
		"cP8uJJpHzJ7ufTPjYuWoVF4s.3MUdOR9iwcO.6E3uCHX1waqypjey458NKGE9O7lnWpV/" +
		"qd2tg1:19029::::::")

	suite.mockExecutor.On("Exec", "getent", "shadow", "hacluster").Return(
		shadow, nil)

	verifyPasswordGatherer := NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "password",
			Argument: "hacluster:linux",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(factRequests)

	expectedResults := []Fact{
		{
			Name:    "hacluster",
			Value:   true,
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PasswordTestSuite) TestPasswordGatherNotEqual() {
	shadow := []byte("hacluster:$6$WFEgSAefduOyvLCN$MprO90En7b/" +
		"cP8uJJpHzJ7ufTPjYuWoVF4s.3MUdOR9iwcO.6E3uCHX1waqypjey458NKGE9O7lnWpV" +
		"/qd2tg1:19029::::::")

	suite.mockExecutor.On("Exec", "getent", "shadow", "hacluster").Return(
		shadow, nil)

	verifyPasswordGatherer := NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "password",
			Argument: "hacluster:linux",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(factRequests)

	expectedResults := []Fact{
		{
			Name:    "hacluster",
			Value:   false,
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PasswordTestSuite) TestPasswordGatherWrongArguments() {
	verifyPasswordGatherer := &VerifyPasswordGatherer{} // nolint

	factRequests := []FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "password",
			Argument: "linux",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(factRequests)

	expectedResults := []Fact{}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
