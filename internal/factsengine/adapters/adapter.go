package adapters

type Adapter interface {
	Unsubscribe() error
	Listen(agentID string, handle func([]byte) error) error
	Publish(facts []byte) error
}
