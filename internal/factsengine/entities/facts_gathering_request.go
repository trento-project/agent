package entities

import (
	"github.com/trento-project/fabriziosestito/golang/pkg/events"
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

// func FactsGatheringRequestedFromEvent(event *contracts.FactsGatheringRequestedV1) FactsGatheringRequested {
// 	agentFacts := []AgentFacts{}
// 	for _, eventAgentFact := range event.Facts {
// 		facts := []FactDefinition{}
// 		for _, eventFact := range eventAgentFact.Facts {
// 			fact := FactDefinition{
// 				Argument: eventFact.Argument,
// 				CheckID:  eventFact.CheckId,
// 				Gatherer: eventFact.Gatherer,
// 				Name:     eventFact.Name,
// 			}
// 			facts = append(facts, fact)
// 		}
// 		agentFact := AgentFacts{
// 			AgentID: eventAgentFact.AgentId,
// 			Facts:   facts,
// 		}
// 		agentFacts = append(agentFacts, agentFact)
// 	}

// 	return FactsGatheringRequested{
// 		ExecutionID: event.ExecutionId,
// 		Facts:       agentFacts,
// 	}
// }

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
		Facts:       agentFacts,
	}, nil
}
