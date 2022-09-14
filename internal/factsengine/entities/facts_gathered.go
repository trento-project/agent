// nolint:nosnakecase
package entities

import (
	"bytes"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/trento-project/contracts/golang/pkg/events"
)

const (
	eventSource = "https://github.com/trento-project/agent"
)

type Error struct {
	Message string
	Type    string
}

type Fact struct {
	Name    string
	CheckID string
	Value   interface{}
	Error   *Error
}

type FactsGathered struct {
	AgentID       string
	ExecutionID   string
	FactsGathered []Fact
}

func NewFactGatheredWithRequest(factReq FactRequest, value interface{}) Fact {
	return Fact{
		Name:    factReq.Name,
		CheckID: factReq.CheckID,
		Value:   value,
		Error:   nil,
	}
}

func factGatheredItemToEvent(fact Fact) *events.Fact {
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
			eventFact = &events.Fact{
				CheckId: fact.CheckID,
				Name:    fact.Name,
				Value: &events.Fact_TextValue{
					TextValue: value,
				},
			}
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

func FactsGatheredToEvent(gatheredFacts FactsGathered) ([]byte, error) {
	facts := []*events.Fact{}
	for _, fact := range gatheredFacts.FactsGathered {
		facts = append(facts, factGatheredItemToEvent(fact))
	}

	event := events.FactsGathered{
		AgentId:       gatheredFacts.AgentID,
		ExecutionId:   gatheredFacts.ExecutionID,
		FactsGathered: facts,
	}

	eventBytes, err := events.ToEvent(&event, eventSource, uuid.New().String())
	if err != nil {
		return nil, errors.Wrap(err, "error creating event")
	}

	return eventBytes, nil
}

func PrettifyEvent(data interface{}) (string, error) {
	jsonResult, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "Error building the response")
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonResult, "", "  "); err != nil {
		return "", errors.Wrap(err, "Error indenting the json data")
	}

	return prettyJSON.String(), nil
}
