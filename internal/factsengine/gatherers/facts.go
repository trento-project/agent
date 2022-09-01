package gatherers

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

type FactsResult struct {
	ExecutionID string `json:"execution_id"`
	AgentID     string `json:"agent_id"`
	Facts       []Fact `json:"facts_gathered"`
}

type Fact struct {
	Name    string      `json:"name"`
	Value   interface{} `json:"value"`
	CheckID string      `json:"check_id"`
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
