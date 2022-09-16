package entities

const (
	FactsGathererdEventSource = "https://github.com/trento-project/agent"
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
