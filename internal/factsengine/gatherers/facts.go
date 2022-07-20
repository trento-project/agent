package gatherers

type FactsResult struct {
	ExecutionID string `json:"execution_id"`
	Facts       []Fact `json:"facts"`
}

type Fact struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
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
}
