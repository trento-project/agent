package entities

import (
	"github.com/trento-project/contracts/pkg/events"
)

type FactsGatheringRequestedTarget struct {
	AgentID      string
	FactRequests []FactRequest
}

type FactRequest struct {
	Argument string
	CheckID  string
	Gatherer string
	Name     string
}

type FactsGatheringRequested struct {
	ExecutionID string
	Targets     []FactsGatheringRequestedTarget
}

type GroupedByGathererRequestedTarget struct {
	FactRequests map[string][]FactRequest
}

func FactsGatheringRequestedFromEvent(event []byte) (*FactsGatheringRequested, error) {
	var factsGatheringRequestedEvent events.FactsGatheringRequested

	err := events.FromEvent(event, &factsGatheringRequestedEvent)
	if err != nil {
		return nil, err
	}

	targets := []FactsGatheringRequestedTarget{}
	for _, eventAgentFact := range factsGatheringRequestedEvent.GetTargets() {
		factRequests := []FactRequest{}
		for _, eventFact := range eventAgentFact.GetFactRequests() {
			fact := FactRequest{
				Argument: eventFact.GetArgument(),
				CheckID:  eventFact.GetCheckId(),
				Gatherer: eventFact.GetGatherer(),
				Name:     eventFact.GetName(),
			}
			factRequests = append(factRequests, fact)
		}
		target := FactsGatheringRequestedTarget{
			AgentID:      eventAgentFact.GetAgentId(),
			FactRequests: factRequests,
		}
		targets = append(targets, target)
	}

	return &FactsGatheringRequested{
		ExecutionID: factsGatheringRequestedEvent.ExecutionId,
		Targets:     targets,
	}, nil
}
