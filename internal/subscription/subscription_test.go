package subscription

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/trento-project/agent/internal/subscription/mocks"
)

type SubscriptionTestSuite struct {
	suite.Suite
}

func TestSubscriptionTestSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionTestSuite))
}

func mockSUSEConnect() *exec.Cmd {
	return exec.Command("echo", `[{"identifier":"SLES_SAP","version":"15.2","arch":"x86_64",
    "status":"Registered","name":"SUSE Employee subscription for SUSE Linux Enterprise Server for SAP Applications",
    "regcode":"my-code","starts_at":"2019-03-20 09:55:32 UTC",
    "expires_at":"2024-03-20 09:55:32 UTC","subscription_status":"ACTIVE","type":"internal"},
    {"identifier":"sle-module-public-cloud","version":"15.2",
    "arch":"x86_64","status":"Registered"}]`)
}

func mockSUSEConnectErr() *exec.Cmd {
	return exec.Command("error")
}

func (suite *SubscriptionTestSuite) TestNewSubscriptions() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "SUSEConnect", "-s").Return(
		mockSUSEConnect(),
	)

	subs, err := NewSubscriptions()

	expectedSubs := Subscriptions{
		&Subscription{
			Identifier:         "SLES_SAP",
			Version:            "15.2",
			Arch:               "x86_64",
			Status:             "Registered",
			StartsAt:           "2019-03-20 09:55:32 UTC",
			ExpiresAt:          "2024-03-20 09:55:32 UTC",
			SubscriptionStatus: "ACTIVE",
			Type:               "internal",
		},
		&Subscription{ //nolint
			Identifier: "sle-module-public-cloud",
			Version:    "15.2",
			Arch:       "x86_64",
			Status:     "Registered",
		},
	}

	suite.NoError(err)
	suite.ElementsMatch(expectedSubs, subs)
}

func (suite *SubscriptionTestSuite) TestNewSubscriptionsErr() {
	mockCommand := new(mocks.CustomCommand)

	customExecCommand = mockCommand.Execute

	mockCommand.On("Execute", "SUSEConnect", "-s").Return(
		mockSUSEConnectErr(),
	)

	subs, err := NewSubscriptions()

	suite.Equal(Subscriptions(nil), subs)
	suite.EqualError(err, "exec: \"error\": executable file not found in $PATH")
}
