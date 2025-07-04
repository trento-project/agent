package main

// go build -o /usr/etc/trento/dummy ./plugin_examples/dummy/dummy.go

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/hashicorp/go-plugin"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/factsengine/plugininterface"
)

type dummyGatherer struct {
}

func (s dummyGatherer) Gather(_ context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	slog.Info("Starting dummy plugin facts gathering process")

	for _, factReq := range factsRequests {
		value := rand.Int() // nolint
		fact := entities.NewFactGatheredWithRequest(factReq, &entities.FactValueString{Value: fmt.Sprint(value)})
		facts = append(facts, fact)
	}

	slog.Info("Requested dummy plugin facts gathered")
	return facts, nil
}

func main() {
	d := &dummyGatherer{}

	handshakeConfig := plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "TRENTO_PLUGIN",
		MagicCookieValue: "gatherer",
	}

	var pluginMap = map[string]plugin.Plugin{
		"gatherer": &plugininterface.GathererPlugin{Impl: d},
	}

	plugin.Serve(&plugin.ServeConfig{ // nolint
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
