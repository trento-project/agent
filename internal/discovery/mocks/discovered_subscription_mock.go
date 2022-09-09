package mocks

import (
	"encoding/json"
	"io"
	"os"

	"github.com/trento-project/agent/internal/subscription"
)

func NewDiscoveredSubscriptionsMock() subscription.Subscriptions {
	var subs subscription.Subscriptions

	jsonFile, err := os.Open("./test/fixtures/discovery/subscriptions/subscriptions_discovery.json")
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
