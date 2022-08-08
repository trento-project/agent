package adapters

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
)

const (
	gatherFactsExchanage string = "gather_facts"
	factsExchanage       string = "facts"
)

type RabbitMQAdapter struct {
	consumer  rabbitmq.Consumer
	publisher *rabbitmq.Publisher
}

func NewRabbitMQAdapter(factsEngineService string) (*RabbitMQAdapter, error) {
	consumer, err := rabbitmq.NewConsumer(
		factsEngineService,
		rabbitmq.Config{}, //nolint
		rabbitmq.WithConsumerOptionsLogging,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating the rabbitmq consumer")
	}

	publisher, err := rabbitmq.NewPublisher(
		factsEngineService,
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

func (r *RabbitMQAdapter) Listen(agentID string, handle func(string, []byte) error) error {
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
		agentID,
		[]string{agentID},
		rabbitmq.WithConsumeOptionsQueueDurable,
		rabbitmq.WithConsumeOptionsQueueAutoDelete,
		rabbitmq.WithConsumeOptionsBindingExchangeName(gatherFactsExchanage),
		rabbitmq.WithConsumeOptionsBindingExchangeDurable,
		rabbitmq.WithConsumeOptionsBindingExchangeAutoDelete,
	)
}

func (r *RabbitMQAdapter) Publish(facts []byte, contentType string) error {
	return r.publisher.Publish(
		facts,
		[]string{""},
		rabbitmq.WithPublishOptionsContentType(contentType),
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsExchange(factsExchanage),
	)
}
