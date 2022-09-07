package entities

import (
	contracts "github.com/trento-project/contracts/go/pkg/gen/entities"
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
	Facts       []AgentFacts
}

type GroupedByGathererAgentFacts struct {
	Facts map[string][]FactDefinition
}

func FactsGatheringRequestedFromEvent(event *contracts.FactsGatheringRequestedV1) FactsGatheringRequested {
	agentFacts := []AgentFacts{}
	for _, eventAgentFact := range event.Facts {
		facts := []FactDefinition{}
		for _, eventFact := range eventAgentFact.Facts {
			fact := FactDefinition{
				Argument: eventFact.Argument,
				CheckID:  eventFact.CheckId,
				Gatherer: eventFact.Gatherer,
				Name:     eventFact.Name,
			}
			facts = append(facts, fact)
		}
		agentFact := AgentFacts{
			AgentID: eventAgentFact.AgentId,
			Facts:   facts,
		}
		agentFacts = append(agentFacts, agentFact)
	}

	return FactsGatheringRequested{
		ExecutionID: event.ExecutionId,
		Facts:       agentFacts,
	}
}
