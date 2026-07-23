// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package main

// go build -o /usr/etc/trento/dummy ./plugin_examples/dummy/dummy.go

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math"
	"math/big"
	"strconv"

	"github.com/hashicorp/go-plugin"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/factsengine/plugininterface"
)

type dummyGatherer struct {
}

func (s dummyGatherer) Gather(_ context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := make([]entities.Fact, 0, len(factsRequests))

	slog.Info("Starting dummy plugin facts gathering process")

	for _, factReq := range factsRequests {
		value, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			return nil, err
		}

		fact := entities.NewFactGatheredWithRequest(
			factReq,
			&entities.FactValueString{Value: strconv.FormatInt(value.Int64(), 10)},
		)

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

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
