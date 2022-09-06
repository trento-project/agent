package entities

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	contracts "github.com/trento-project/contracts/go/pkg/gen/entities"
)

type FactGatheringError struct {
	Message string
	Type    string
}

type FactsGatheredItem struct {
	Name    string
	CheckID string
	Value   interface{}
	Error   *FactGatheringError
}

type FactsGathered struct {
	AgentID       string
	ExecutionID   string
	FactsGathered []FactsGatheredItem
}

func (e FactGatheringError) Error() string {
	return fmt.Sprintf("fact gathering error: type: %s - %s", e.Type, e.Message)
}

func (e FactGatheringError) Wrap(msg string) FactGatheringError {
	return FactGatheringError{
		Message: fmt.Sprintf("%s: %v", e.Message, msg),
		Type:    e.Type,
	}
}

func NewFactGatheredWithRequest(req FactRequest, value interface{}) FactsGatheredItem {
	return FactsGatheredItem{
		Name:    req.Name,
		CheckID: req.CheckID,
		Value:   value,
		Error:   nil,
	}
}

func NewFactGatheredWithError(req FactRequest, err *FactGatheringError) FactsGatheredItem {
	return FactsGatheredItem{
		Name:    req.Name,
		CheckID: req.CheckID,
		Value:   nil,
		Error:   err,
	}
}

func NewFactsGatheredListWithError(reqs []FactRequest, err *FactGatheringError) []FactsGatheredItem {
	factsWithErrors := []FactsGatheredItem{}
	for _, req := range reqs {
		factsWithErrors = append(factsWithErrors, NewFactGatheredWithError(req, err))
	}

	return factsWithErrors
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
