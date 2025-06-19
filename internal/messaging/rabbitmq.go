package messaging

import (
	"log/slog"

	"github.com/pkg/errors"
	"github.com/wagslane/go-rabbitmq"
)

type RabbitMQAdapter struct {
	conn      *rabbitmq.Conn
	consumer  *rabbitmq.Consumer
	publisher *rabbitmq.Publisher
	exchange  string
}

func NewRabbitMQAdapter(
	connectionURI string,
	queue,
	exchange,
	routingKey string,
) (*RabbitMQAdapter, error) {
	conn, err := rabbitmq.NewConn(
		connectionURI,
		rabbitmq.WithConnectionOptionsLogging,
	)

	if err != nil {
		return nil, errors.Wrap(err, "could not create rabbitmq connection")
	}

	consumer, err := rabbitmq.NewConsumer(
		conn,
		queue,
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchange),
		rabbitmq.WithConsumerOptionsExchangeKind("topic"),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeDurable,
		rabbitmq.WithConsumerOptionsQueueDurable,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create consumer")
	}

	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create publisher")
	}

	return &RabbitMQAdapter{
		consumer:  consumer,
		publisher: publisher,
		conn:      conn,
		exchange:  exchange,
	}, nil

}

func (r *RabbitMQAdapter) Unsubscribe() error {
	r.consumer.Close()
	r.publisher.Close()
	return r.conn.Close()
}

func (r *RabbitMQAdapter) Listen(
	handle func(contentType string, message []byte) error,
) error {
	// Cancelation is handled internally on the library with Consumer closing, safe to just spawn
	go func() {
		err := r.consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
			// TODO: Handle different kind of errors returning some sort of metadata
			// so the applied action is potentially changed
			err := handle(d.ContentType, d.Body)
			if err != nil {
				slog.Error("error handling message", "error", err)
				return rabbitmq.NackDiscard
			}

			return rabbitmq.Ack
		})
		if err != nil {
			slog.Error("consumer run failed", "error", err)
		}
	}()

	return nil
}

func (r *RabbitMQAdapter) Publish(
	routingKey,
	contentType string,
	message []byte,
) error {
	return r.publisher.Publish(
		message,
		[]string{routingKey},
		rabbitmq.WithPublishOptionsContentType(contentType),
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsExchange(r.exchange),
	)
}
