package adapters

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
)

type RabbitMQAdapter struct {
	conn      *rabbitmq.Conn
	consumer  *rabbitmq.Consumer
	publisher *rabbitmq.Publisher
}

func NewRabbitMQAdapter(url string) (*RabbitMQAdapter, error) {
	conn, err := rabbitmq.NewConn(
		url,
		rabbitmq.WithConnectionOptionsLogging,
	)

	if err != nil {
		return nil, errors.Wrap(err, "could not create rabbitmq connection")
	}

	return &RabbitMQAdapter{
		consumer:  nil,
		publisher: nil,
		conn:      conn,
	}, nil
}

func (r *RabbitMQAdapter) Unsubscribe() error {
	if r.consumer != nil {
		r.consumer.Close()
	}
	if r.publisher != nil {
		r.publisher.Close()
	}
	return r.conn.Close()
}

func (r *RabbitMQAdapter) Listen(
	queue,
	exchange,
	routingKey string,
	handle func(contentType string, message []byte) error,
) error {
	consumer, err := rabbitmq.NewConsumer(
		r.conn,
		queue,
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchange),
		rabbitmq.WithConsumerOptionsExchangeKind("topic"),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeDurable,
		rabbitmq.WithConsumerOptionsQueueDurable,
	)
	if err != nil {
		return err
	}

	r.consumer = consumer

	// Cancelation is handled internally on the library with Consumer closing, safe to just spawn
	go func() {
		err := consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
			// TODO: Handle different kind of errors returning some sort of metadata
			// so the applied action is potentially changed
			err := handle(d.ContentType, d.Body)
			if err != nil {
				log.Errorf("error handling message: %s", err)
				return rabbitmq.NackDiscard
			}

			return rabbitmq.Ack
		})
		if err != nil {
			log.Error(err)
		}
	}()

	return nil
}

func (r *RabbitMQAdapter) Publish(
	exchange,
	routingKey,
	contentType string, message []byte,
) error {
	publisher, err := rabbitmq.NewPublisher(
		r.conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName("events"),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	if err != nil {
		return err
	}

	r.publisher = publisher

	return publisher.Publish(
		message,
		[]string{routingKey},
		rabbitmq.WithPublishOptionsContentType(contentType),
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsExchange(exchange),
	)
}
