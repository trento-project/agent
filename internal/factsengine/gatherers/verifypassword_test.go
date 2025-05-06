package gatherers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
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

	suite.mockExecutor.On("ExecContext", mock.Anything, "/usr/bin/getent", "shadow", "hacluster").Return(
		shadow, nil)

	verifyPasswordGatherer := gatherers.NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(context.Background(), factRequests)

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

	suite.mockExecutor.On("ExecContext", mock.Anything, "/usr/bin/getent", "shadow", "hacluster").Return(
		shadow, nil)

	verifyPasswordGatherer := gatherers.NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(context.Background(), factRequests)

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

func (suite *PasswordTestSuite) TestPasswordGatherBloquedPassword() {
	suite.mockExecutor.
		On("ExecContext", mock.Anything, "/usr/bin/getent", "shadow", "hacluster").
		Return([]byte("hacluster:!:19029::::::"), nil).
		Once().
		On("ExecContext", mock.Anything, "/usr/bin/getent", "shadow", "hacluster").
		Return([]byte("hacluster:!$6$WFEgSAefduOyvLCN$MprO90E:19029::::::"), nil).
		Once().
		On("ExecContext", mock.Anything, "/usr/bin/getent", "shadow", "hacluster").
		Return([]byte("hacluster:*:19029::::::"), nil)

	verifyPasswordGatherer := gatherers.NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check1",
		},
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check2",
		},
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check3",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "hacluster",
			CheckID: "check2",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "password authentication blocked for user: hacluster",
				Type:    "verify-password-password-blocked",
			},
		},
		{
			Name:    "hacluster",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "password authentication blocked for user: hacluster",
				Type:    "verify-password-password-blocked",
			},
		},
		{
			Name:    "hacluster",
			CheckID: "check3",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "password authentication blocked for user: hacluster",
				Type:    "verify-password-password-blocked",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PasswordTestSuite) TestPasswordGatherNoPassword() {
	shadow := []byte("hacluster::19029::::::")

	suite.mockExecutor.On("ExecContext", mock.Anything, "/usr/bin/getent", "shadow", "hacluster").Return(
		shadow, nil)

	verifyPasswordGatherer := gatherers.NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "hacluster",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "password not set for user: hacluster",
				Type:    "verify-password-password-not-set",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PasswordTestSuite) TestPasswordGatherDifferentEncType() {
	shadow := []byte("hacluster:$5$WFEgPAefduOyvLCN$MprO90En7b/" +
		"cP8uJJpHzJ7ufTPjYuWoVF4s.3MUdOR9iwcO.6E3uCHX1waqypjey458NKGE9O7lnWpV/" +
		"qd2tg1:19029::::::")

	suite.mockExecutor.On("ExecContext", mock.Anything, "/usr/bin/getent", "shadow", "hacluster").Return(
		shadow, nil)

	verifyPasswordGatherer := gatherers.NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "hacluster",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "error while verifying the password for user: hacluster: invalid magic prefix",
				Type:    "verify-password-crypt-error",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PasswordTestSuite) TestPasswordGatherInvalidShadowOutput() {
	shadow := []byte("hacluster:hash:")

	suite.mockExecutor.On("ExecContext", mock.Anything, "/usr/bin/getent", "shadow", "hacluster").Return(
		shadow, nil)

	verifyPasswordGatherer := gatherers.NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "hacluster",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "error getting shadow output: shadow output does not have 9 fields: hacluster:hash:",
				Type:    "verify-password-shadow-error",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PasswordTestSuite) TestPasswordGatherWrongArguments() {
	verifyPasswordGatherer := gatherers.NewVerifyPasswordGatherer(suite.mockExecutor)

	factRequests := []entities.FactRequest{
		{
			Name:     "pepito",
			Gatherer: "verify_password",
			Argument: "pepito",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(context.Background(), factRequests)

	expectedResults := []entities.Fact{
		{
			Name:    "pepito",
			CheckID: "check1",
			Value:   nil,
			Error: &entities.FactGatheringError{
				Message: "requested user is not whitelisted for password check: pepito",
				Type:    "verify-password-invalid-username",
			},
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedResults, factResults)
}

func (suite *PasswordTestSuite) TestPasswordGatherContextCancelled() {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	verifyPasswordGatherer := gatherers.NewVerifyPasswordGatherer(utils.Executor{})
	factRequests := []entities.FactRequest{
		{
			Name:     "hacluster",
			Gatherer: "verify_password",
			Argument: "hacluster",
			CheckID:  "check1",
		},
	}

	factResults, err := verifyPasswordGatherer.Gather(ctx, factRequests)

	suite.Error(err)
	suite.Empty(factResults)
}
