package facts

// TODO: Most probably we need different types of facts
type Fact struct {
	Name  string      `json:"name"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type FactsRequest struct {
	Name string   `json:"name"`
	Keys []string `json:"keys"`
}
