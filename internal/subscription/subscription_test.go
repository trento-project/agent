package subscription

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"

	"github.com/trento-project/agent/internal/utils/mocks"
)

type SubscriptionTestSuite struct {
	suite.Suite
}

func TestSubscriptionTestSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionTestSuite))
}

func (suite *SubscriptionTestSuite) TestNewSubscriptions() {
	mockCommand := new(mocks.CommandExecutor)

	subsOutput := []byte(`[{"identifier":"SLES_SAP","version":"15.2","arch":"x86_64",
    "status":"Registered","name":"SUSE Employee subscription for SUSE Linux Enterprise Server for SAP Applications",
    "regcode":"my-code","starts_at":"2019-03-20 09:55:32 UTC",
    "expires_at":"2024-03-20 09:55:32 UTC","subscription_status":"ACTIVE","type":"internal"},
    {"identifier":"sle-module-public-cloud","version":"15.2",
    "arch":"x86_64","status":"Registered"}]`)

	mockCommand.On("Exec", "SUSEConnect", "-s").Return(
		subsOutput, nil,
	)

	subs, err := NewSubscriptions(mockCommand)

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
	mockCommand := new(mocks.CommandExecutor)

	mockCommand.On("Exec", "SUSEConnect", "-s").Return(
		nil, errors.New("some error"),
	)

	subs, err := NewSubscriptions(mockCommand)

	suite.Equal(Subscriptions(nil), subs)
	suite.EqualError(err, "some error")
}
