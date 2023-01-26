package entities

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	FactsGathererdEventSource = "https://github.com/trento-project/agent"
)

type FactGatheringError struct {
	Message string
	Type    string
}

type Fact struct {
	Name    string
	CheckID string
	Value   FactValue
	Error   *FactGatheringError
}

type FactsGathered struct {
	AgentID       string
	ExecutionID   string
	FactsGathered []Fact
	GroupID       string
}

func (e *FactGatheringError) Error() string {
	return fmt.Sprintf("fact gathering error: %s - %s", e.Type, e.Message)
}

func (e *FactGatheringError) Wrap(msg string) *FactGatheringError {
	return &FactGatheringError{
		Message: fmt.Sprintf("%s: %v", e.Message, msg),
		Type:    e.Type,
	}
}

func (e *Fact) Prettify() (string, error) {
	prettifiedValue, err := Prettify(e.Value)
	if err != nil {
		return "", errors.Wrap(err, "Error prettifying fact value data")
	}

	result := fmt.Sprintf("Name: %s\nCheck ID: %s\n\nValue:\n\n%s", e.Name, e.CheckID, prettifiedValue)
	return result, nil
}

func NewFactGatheredWithRequest(factReq FactRequest, value FactValue) Fact {
	return Fact{
		Name:    factReq.Name,
		CheckID: factReq.CheckID,
		Value:   value,
		Error:   nil,
	}
}

func NewFactGatheredWithError(req FactRequest, err *FactGatheringError) Fact {
	return Fact{
		Name:    req.Name,
		CheckID: req.CheckID,
		Value:   nil,
		Error:   err,
	}
}

func NewFactsGatheredListWithError(reqs []FactRequest, err *FactGatheringError) []Fact {
	factsWithErrors := []Fact{}
	for _, req := range reqs {
		factsWithErrors = append(factsWithErrors, NewFactGatheredWithError(req, err))
	}

	return factsWithErrors
}
