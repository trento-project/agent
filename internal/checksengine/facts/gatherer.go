package facts

type FactGatherer interface {
	Gather(keys []string) ([]*Fact, error)
}
