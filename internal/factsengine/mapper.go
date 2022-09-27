// nolint:nosnakecase
package factsengine

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/contracts/go/pkg/events"
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
		Targets:     targets,
	}, nil
}

func factGatheredItemToEvent(fact entities.Fact) *events.Fact {
	var eventFact *events.Fact

	if fact.Error != nil {
		eventFact = &events.Fact{
			CheckId: fact.CheckID,
			Name:    fact.Name,
			Value: &events.Fact_ErrorValue{
				ErrorValue: &events.FactError{
					Message: fact.Error.Message,
					Type:    fact.Error.Type,
				},
			},
		}
	} else {
		switch value := fact.Value.(type) {
		case string:
			eventFact = parseString(value, fact)
		case int:
			eventFact = &events.Fact{
				CheckId: fact.CheckID,
				Name:    fact.Name,
				Value: &events.Fact_NumericValue{
					NumericValue: float32(value),
				},
			}
		case nil:
			eventFact = &events.Fact{
				CheckId: fact.CheckID,
				Name:    fact.Name,
				Value: &events.Fact_ErrorValue{
					ErrorValue: &events.FactError{
						Message: "null value",
						Type:    "null_value",
					},
				},
			}
		default:
			eventFact = &events.Fact{
				CheckId: fact.CheckID,
				Name:    fact.Name,
				Value: &events.Fact_ErrorValue{
					ErrorValue: &events.FactError{
						Message: "unknown value type",
						Type:    "unknown_value_type",
					},
				},
			}
		}

	}

	return eventFact
}

func parseString(value string, fact entities.Fact) *events.Fact {
	if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
		return &events.Fact{
			CheckId: fact.CheckID,
			Name:    fact.Name,
			Value: &events.Fact_NumericValue{
				NumericValue: float32(floatValue),
			},
		}
	}

	return &events.Fact{
		CheckId: fact.CheckID,
		Name:    fact.Name,
		Value: &events.Fact_TextValue{
			TextValue: value,
		},
	}
}

func FactsGatheredToEvent(gatheredFacts entities.FactsGathered) ([]byte, error) {
	facts := []*events.Fact{}
	for _, fact := range gatheredFacts.FactsGathered {
		facts = append(facts, factGatheredItemToEvent(fact))
	}

	event := events.FactsGathered{
		AgentId:       gatheredFacts.AgentID,
		ExecutionId:   gatheredFacts.ExecutionID,
		FactsGathered: facts,
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
