package adapters

//go:generate mockery --name=Adapter

type Adapter interface {
	Unsubscribe() error
	Listen(agentID string, handle func(string, []byte) error) error
	Publish(facts []byte, contentType string) error
}
