package adapters

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
)

type RabbitMQAdapter struct {
	consumer  rabbitmq.Consumer
	publisher *rabbitmq.Publisher
}

func NewRabbitMQAdapter(url string) (*RabbitMQAdapter, error) {
	consumer, err := rabbitmq.NewConsumer(
		url,
		rabbitmq.Config{}, //nolint
		rabbitmq.WithConsumerOptionsLogging,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating the rabbitmq consumer")
	}

	publisher, err := rabbitmq.NewPublisher(
		url,
		rabbitmq.Config{}, //nolint
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating the rabbitmq publisher")
	}

	return &RabbitMQAdapter{
		consumer:  consumer,
		publisher: publisher,
	}, nil
}

func (r *RabbitMQAdapter) Unsubscribe() error {
	if err := r.consumer.Close(); err != nil {
		return errors.Wrap(err, "Error closing the rabbitmq consumer")
	}

	if err := r.publisher.Close(); err != nil {
		return errors.Wrap(err, "Error closing the rabbitmq publisher")
	}

	return nil
}

func (r *RabbitMQAdapter) Listen(
	queue, exchange string, handle func(contentType string, message []byte) error) error {

	return r.consumer.StartConsuming(
		func(d rabbitmq.Delivery) rabbitmq.Action {
			// TODO: Handle different kind of errors returning some sort of metadata
			// so the applied action is potentially changed
			err := handle(d.ContentType, d.Body)
			if err != nil {
				log.Errorf("error handling message: %s", err)
				return rabbitmq.NackDiscard
			}

			return rabbitmq.Ack
		},
		queue,
		[]string{queue},
		rabbitmq.WithConsumeOptionsQueueDurable,
		rabbitmq.WithConsumeOptionsQueueAutoDelete,
		rabbitmq.WithConsumeOptionsBindingExchangeName(exchange),
		rabbitmq.WithConsumeOptionsBindingExchangeDurable,
		rabbitmq.WithConsumeOptionsBindingExchangeAutoDelete,
	)
}

func (r *RabbitMQAdapter) Publish(exchange, contentType string, message []byte) error {
	return r.publisher.Publish(
		message,
		[]string{""},
		rabbitmq.WithPublishOptionsContentType(contentType),
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsExchange(exchange),
	)
}
