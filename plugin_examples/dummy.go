package main

// go build -o /usr/etc/trento/dummy ./plugin_examples/dummy.go

import (
	"fmt"
	"math/rand"

	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/agent/internal/factsengine/plugininterface"
)

type dummyGatherer struct {
}

func (s dummyGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting dummy plugin facts gathering process")

	for _, factReq := range factsRequests {
		value := rand.Int() // nolint
		fact := entities.NewFactGatheredWithRequest(factReq, fmt.Sprint(value))
		facts = append(facts, fact)
	}

	log.Infof("Requested dummy plugin facts gathered")
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
