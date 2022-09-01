package gatherers

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	contracts "github.com/trento-project/contracts/go/pkg/gen/entities"
)

type FactsResult struct {
	ExecutionID string
	AgentID     string
	Facts       []Fact
}

type Fact struct {
	Name    string
	Value   interface{}
	CheckID string
}

type FactsRequest struct {
	ExecutionID string        `json:"execution_id"`
	Facts       []FactRequest `json:"facts"`
}

type GroupedFactsRequest struct {
	ExecutionID string
	Facts       map[string][]FactRequest
}

type FactRequest struct {
	Name     string `json:"name"`
	Gatherer string `json:"gatherer"`
	Argument string `json:"argument"`
	CheckID  string `json:"check_id"`
}

func NewFactWithRequest(req FactRequest, value interface{}) Fact {
	return Fact{
		Name:    req.Name,
		CheckID: req.CheckID,
		Value:   value,
	}
}

func FactsGatheredToEvent(gatheredFacts FactsResult) contracts.FactsGatheredV1 {
	facts := []*contracts.FactsGatheredItems{}
	for _, fact := range gatheredFacts.Facts {
		eventFact := &contracts.FactsGatheredItems{
			CheckId: fact.CheckID,
			Error:   nil, // TODO: Set error once is it defined in the code
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

func PrettifyFactResult(fact Fact) (string, error) {
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
