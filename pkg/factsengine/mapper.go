// nolint:nosnakecase
package factsengine

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/contracts/go/pkg/events"
	"google.golang.org/protobuf/types/known/structpb"
)

func FactsGatheringRequestedFromEvent(event []byte) (*entities.FactsGatheringRequested, error) {
	var factsGatheringRequestedEvent events.FactsGatheringRequested

	err := events.FromEvent(event, &factsGatheringRequestedEvent)
	if err != nil {
		return nil, err
	}

	targets := []entities.FactsGatheringRequestedTarget{}
	for _, eventAgentFact := range factsGatheringRequestedEvent.GetTargets() {
		factRequests := []entities.FactRequest{}
		for _, eventFact := range eventAgentFact.GetFactRequests() {
			fact := entities.FactRequest{
				Argument: eventFact.GetArgument(),
				CheckID:  eventFact.GetCheckId(),
				Gatherer: eventFact.GetGatherer(),
				Name:     eventFact.GetName(),
			}
			factRequests = append(factRequests, fact)
		}
		target := entities.FactsGatheringRequestedTarget{
			AgentID:      eventAgentFact.GetAgentId(),
			FactRequests: factRequests,
		}
		targets = append(targets, target)
	}

	return &entities.FactsGatheringRequested{
		ExecutionID: factsGatheringRequestedEvent.ExecutionId,
		GroupID:     factsGatheringRequestedEvent.GroupId,
		Targets:     targets,
	}, nil
}

func factGatheredItemToEvent(fact entities.Fact) (*events.Fact, error) {
	eventFact := &events.Fact{
		CheckId: fact.CheckID,
		Name:    fact.Name,
	}

	if fact.Error != nil {
		eventFact.FactValue = &events.Fact_ErrorValue{
			ErrorValue: &events.FactError{
				Message: fact.Error.Message,
				Type:    fact.Error.Type,
			},
		}
	} else {
		value, err := structpb.NewValue(fact.Value.AsInterface())

		if err != nil {
			return nil, err
		}

		eventFact.FactValue = &events.Fact_Value{
			Value: value,
		}
	}

	return eventFact, nil
}

func FactsGatheredToEvent(gatheredFacts entities.FactsGathered) ([]byte, error) {
	facts := []*events.Fact{}
	for _, fact := range gatheredFacts.FactsGathered {
		eventFact, err := factGatheredItemToEvent(fact)
		if err != nil {
			return nil, err
		}
		facts = append(facts, eventFact)
	}

	event := events.FactsGathered{
		AgentId:       gatheredFacts.AgentID,
		ExecutionId:   gatheredFacts.ExecutionID,
		FactsGathered: facts,
		GroupId:       gatheredFacts.GroupID,
	}

	eventBytes, err := events.ToEvent(
		&event,
		events.WithSource(entities.FactsGathererdEventSource),
		events.WithID(uuid.New().String()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating event")
	}

	return eventBytes, nil
}
