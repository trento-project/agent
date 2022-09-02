package entities

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
