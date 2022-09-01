package adapters

//go:generate mockery --name=Adapter

type Adapter interface {
	Unsubscribe() error
	Listen(agentID string, handle func(contentType string, factsRequest []byte) error) error
	Publish(contentType string, facts []byte) error
}
