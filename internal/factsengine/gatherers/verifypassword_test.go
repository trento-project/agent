package gatherers

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	utilsMocks "github.com/trento-project/agent/pkg/utils/mocks"
)

type PasswordTestSuite struct {
	suite.Suite
	mockExecutor *utilsMocks.CommandExecutor
}

func TestPasswordTestSuite(t *testing.T) {
	suite.Run(t, new(PasswordTestSuite))
}

func (suite *PasswordTestSuite) SetupTest() {
	suite.mockExecutor = new(utilsMocks.CommandExecutor)
}

func (suite *PasswordTestSuite) TestPasswordGatherEqual() {
	shadow := []byte("hacluster:$6$WFEgPAefduOyvLCN$MprO90En7b/" +
		"cP8uJJpHzJ7ufTPjYuWoVF4s.3MUdOR9iwcO.6E3uCHX1waqypjey458NKGE9O7lnWpV/" +
		"qd2tg1:19029::::::")

	suite.mockExecutor.On("Exec", "getent", "shadow", "hacluster").Return(
		shadow, nil)

	verifyPasswordGatherer := NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "hacluster",
			Value:   &entities.FactValueBool{Value: true},
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

	factRequests := []entities.FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "hacluster",
			Value:   &entities.FactValueBool{Value: false},
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PasswordTestSuite) TestPasswordGatherWrongArguments() {
	verifyPasswordGatherer := &VerifyPasswordGatherer{} // nolint

	factRequests := []entities.FactRequest{
		{
			Name:     "pepito",
			Gatherer: "verify_password",
			Argument: "pepito",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "pepito",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "unknown username or not allowed to check: pepito",
				Type:    "verify-password-invalid-username",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}
