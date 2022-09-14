package entities

import (
	"github.com/trento-project/contracts/golang/pkg/events"
)

type AgentFacts struct {
	AgentID string
	Facts   []FactDefinition
}

type FactDefinition struct {
	Argument string
	CheckID  string
	Gatherer string
	Name     string
}

type FactsGatheringRequested struct {
	ExecutionID string
	Agents      []AgentFacts
}

type GroupedByGathererAgentFacts struct {
	Facts map[string][]FactDefinition
}

func FactsGatheringRequestedFromEvent(event []byte) (*FactsGatheringRequested, error) {
	var factsGatheringRequestedEvent events.FactsGatheringRequested

	err := events.FromEvent(event, &factsGatheringRequestedEvent)
	if err != nil {
		return nil, err
	}

	agentFacts := []AgentFacts{}
	for _, eventAgentFact := range factsGatheringRequestedEvent.GetTargets() {
		facts := []FactDefinition{}
		for _, eventFact := range eventAgentFact.GetFactRequests() {
			fact := FactDefinition{
				Argument: eventFact.GetArgument(),
				CheckID:  eventFact.GetCheckId(),
				Gatherer: eventFact.GetGatherer(),
				Name:     eventFact.GetName(),
			}
			facts = append(facts, fact)
		}
		agentFact := AgentFacts{
			AgentID: eventAgentFact.GetAgentId(),
			Facts:   facts,
		}
		agentFacts = append(agentFacts, agentFact)
	}

	return &FactsGatheringRequested{
		ExecutionID: factsGatheringRequestedEvent.ExecutionId,
		Agents:      agentFacts,
	}, nil
}
