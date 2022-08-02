package gatherers

type FactsResult struct {
	ExecutionID string `json:"execution_id"`
	AgentID     string `json:"agent_id"`
	Facts       []Fact `json:"facts"`
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
