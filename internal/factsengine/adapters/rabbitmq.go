package adapters

import (
	"fmt"

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
		rabbitmq.Config{},
		rabbitmq.WithConsumerOptionsLogging,
	)
	if err != nil {
		return nil, fmt.Errorf("Error creating the rabbitmq consumer: %s", err)
	}

	publisher, err := rabbitmq.NewPublisher(
		factsEngineService,
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		return nil, fmt.Errorf("Error creating the rabbitmq publisher: %s", err)
	}

	return &RabbitMQAdapter{
		consumer:  consumer,
		publisher: publisher,
	}, nil
}

func (r *RabbitMQAdapter) Unsubscribe() error {
	if err := r.consumer.Close(); err != nil {
		return fmt.Errorf("Error closing the rabbitmq consumer: %s", err)
	}

	if err := r.publisher.Close(); err != nil {
		return fmt.Errorf("Error closing the rabbitmq publisher: %s", err)
	}

	return nil
}

func (r *RabbitMQAdapter) Listen(agentID string, handle func([]byte) error) error {
	return r.consumer.StartConsuming(
		func(d rabbitmq.Delivery) rabbitmq.Action {
			// TODO: Handle different kind of errors returning some sort of metadata
			// so the applied action is potentially changed
			err := handle(d.Body)
			if err != nil {
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

func (r *RabbitMQAdapter) Publish(facts []byte) error {
	return r.publisher.Publish(
		facts,
		[]string{""},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsExchange(factsExchanage),
	)
}
