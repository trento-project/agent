package gatherers

type FactGatherer interface {
	Gather(factRequests []*FactRequest) ([]*Fact, error)
}
