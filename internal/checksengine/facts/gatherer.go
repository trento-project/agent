package facts

type FactGatherer interface {
	Gather(factRequests []FactRequest) ([]*Fact, error)
}
