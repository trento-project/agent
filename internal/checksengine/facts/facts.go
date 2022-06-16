package facts

// TODO: Most probably we need different types of facts
type Fact struct {
	Name  string      `json:"name"`
	Alias string      `json:"alias"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type FactsRequest struct {
	Type  string        `json:"type"`
	Facts []FactRequest `json:"facts"`
}

type FactRequest struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitempty"`
}
