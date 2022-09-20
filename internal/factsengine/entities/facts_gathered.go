package entities

import "fmt"

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
	Value   interface{}
	Error   *FactGatheringError
}

type FactsGathered struct {
	AgentID       string
	ExecutionID   string
	FactsGathered []Fact
}

func (e *FactGatheringError) Error() string {
	return fmt.Sprintf("fact gathering error: type: %s - %s", e.Type, e.Message)
}

func (e *FactGatheringError) Wrap(msg string) *FactGatheringError {
	return &FactGatheringError{
		Message: fmt.Sprintf("%s: %v", e.Message, msg),
		Type:    e.Type,
	}
}

func NewFactGatheredWithRequest(factReq FactRequest, value interface{}) Fact {
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
