// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package gatherers

import (
	"context"

	"log/slog"

	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/version"
)

const StatusGathererName = "status"

type StatusGatherer struct {
	agentID string
}

func NewStatusGatherer(agentID string) *StatusGatherer {
	return &StatusGatherer{
		agentID: agentID,
	}
}

func (g *StatusGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := make([]entities.Fact, 0, len(factsRequests))

	slog.Info("Starting facts gathering process", "gatherer", StatusGathererName)

	statusValue := &entities.FactValueMap{
		Value: map[string]entities.FactValue{
			"agent_id": &entities.FactValueString{Value: g.agentID},
			"version":  &entities.FactValueString{Value: version.Version},
		},
	}

	for _, requestedFact := range factsRequests {
		fact := entities.NewFactGatheredWithRequest(requestedFact, statusValue)
		facts = append(facts, fact)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	slog.Info("Requested facts gathered", "gatherer", StatusGathererName)

	return facts, nil
}
