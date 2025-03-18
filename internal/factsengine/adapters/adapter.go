package adapters

//go:generate mockery --name=Adapter

type Adapter interface {
	Unsubscribe() error
	Listen(handle func(contentType string, message []byte) error) error
	Publish(routingKey, contentType string, message []byte) error
}
