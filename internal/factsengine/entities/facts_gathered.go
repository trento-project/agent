package entities

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	contracts "github.com/trento-project/contracts/go/pkg/gen/entities"
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

func FactsGatheredToEvent(gatheredFacts FactsGathered) contracts.FactsGatheredV1 {
	facts := []*contracts.FactsGatheredItems{}
	for _, fact := range gatheredFacts.FactsGathered {
		var factGatheringError *contracts.Error
		if fact.Error != nil {
			factGatheringError = &contracts.Error{
				Message: fact.Error.Message,
				Type:    fact.Error.Type,
			}
		}

		eventFact := &contracts.FactsGatheredItems{
			CheckId: fact.CheckID,
			Error:   factGatheringError,
			Name:    fact.Name,
			Value:   fact.Value,
		}
		facts = append(facts, eventFact)
	}

	return contracts.FactsGatheredV1{
		AgentId:       gatheredFacts.AgentID,
		ExecutionId:   gatheredFacts.ExecutionID,
		FactsGathered: facts,
	}
}

func PrettifyFactsGatheredItem(fact FactsGatheredItem) (string, error) {
	jsonResult, err := json.Marshal(fact)
	if err != nil {
		return "", errors.Wrap(err, "Error building the response")
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonResult, "", "  "); err != nil {
		return "", errors.Wrap(err, "Error indenting the json data")
	}

	return prettyJSON.String(), nil
}
