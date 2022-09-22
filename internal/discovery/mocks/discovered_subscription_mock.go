package mocks

import (
	"encoding/json"
	"io"
	"os"

	"github.com/trento-project/agent/internal/subscription"
	"github.com/trento-project/agent/test/helpers"
)

func NewDiscoveredSubscriptionsMock() subscription.Subscriptions {
	var subs subscription.Subscriptions

	jsonFile, err := os.Open(helpers.GetFixtureFile("discovery/subscriptions/subscriptions_discovery.json"))
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &subs)
	if err != nil {
		panic(err)
	}
	return subs
}
