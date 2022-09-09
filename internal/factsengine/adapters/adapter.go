package adapters

//go:generate mockery --name=Adapter

type Adapter interface {
	Unsubscribe() error
	// The exchange parameter of the Listen function defines the binded exchange to the created queue
	Listen(queue, exchange string, handle func(contentType string, message []byte) error) error
	Publish(exchange, contentType string, message []byte) error
}
