package entities

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/trento-project/fabriziosestito/golang/pkg/events"
)

type Error struct {
	Message string
	Type    string
}

type FactsGatheredItem struct {
	Name    string
	CheckID string
	Value   interface{}
	Error   *Error
}

type FactsGathered struct {
	AgentID       string
	ExecutionID   string
	FactsGathered []FactsGatheredItem
}

func NewFactGatheredWithRequest(factDef FactDefinition, value interface{}) FactsGatheredItem {
	return FactsGatheredItem{
		Name:    factDef.Name,
		CheckID: factDef.CheckID,
		Value:   value,
		Error:   nil,
	}
}

// func FactsGatheredToEvent(gatheredFacts FactsGathered) contracts.FactsGatheredV1 {
// 	facts := []*contracts.FactsGatheredItems{}
// 	for _, fact := range gatheredFacts.FactsGathered {
// 		var factGatheringError *contracts.Error
// 		if fact.Error != nil {
// 			factGatheringError = &contracts.Error{
// 				Message: fact.Error.Message,
// 				Type:    fact.Error.Type,
// 			}
// 		}

// 		eventFact := &contracts.FactsGatheredItems{
// 			CheckId: fact.CheckID,
// 			Error:   factGatheringError,
// 			Name:    fact.Name,
// 			Value:   fact.Value,
// 		}

// 		facts = append(facts, eventFact)
// 	}

// 	return contracts.FactsGatheredV1{
// 		AgentId:       gatheredFacts.AgentID,
// 		ExecutionId:   gatheredFacts.ExecutionID,
// 		FactsGathered: facts,
// 	}
// }

func FactGatheredItemToEvent(fact FactsGatheredItem) *events.Fact {
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
		eventFact = &events.Fact{
			CheckId: fact.CheckID,
			Name:    fact.Name,
			Value: &events.Fact_TextValue{
				TextValue: fact.Value.(string),
			},
		}
	}

	return eventFact
}

func FactsGatheredToEvent(gatheredFacts FactsGathered) ([]byte, error) {
	facts := []*events.Fact{}
	for _, fact := range gatheredFacts.FactsGathered {
		facts = append(facts, FactGatheredItemToEvent(fact))
	}

	event := events.FactsGathered{
		AgentId:       gatheredFacts.AgentID,
		ExecutionId:   gatheredFacts.ExecutionID,
		FactsGathered: facts,
	}

	eventBytes, err := events.ToEvent(&event, "source", "b6dd6a2d-a4ae-5497-816a-cfae8a04565f")
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
