package main

// go build -o /usr/etc/trento/dummy ./plugin_examples/dummy.go

import (
	"fmt"
	"math/rand"

	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/plugininterface"
)

type dummyGatherer struct {
}

func (s dummyGatherer) Gather(factsRequests []gatherers.FactRequest) ([]gatherers.Fact, error) {
	var facts []gatherers.Fact
	log.Infof("Starting dummy plugin facts gathering process")

	for i := 0; i < 5; i++ {
		value := rand.Int() // nolint
		fact := gatherers.Fact{
			Name:  fmt.Sprintf("fact_%d", value),
			Value: fmt.Sprint(value),
		}

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
