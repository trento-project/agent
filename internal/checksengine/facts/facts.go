package facts

// TODO: Most probably we need different types of facts
type Fact struct {
	Name     string      `json:"name"`
	Gatherer string      `json:"gatherer"`
	Value    interface{} `json:"value"`
}

type FactRequest struct {
	Name     string `json:"name"`
	Gatherer string `json:"gatherer"`
	Argument string `json:"argument"`
}

type GroupedFactsRequest map[string][]*FactRequest
