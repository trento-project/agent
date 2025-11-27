package discovery

import (
	"context"
	"time"

	"github.com/trento-project/agent/internal/discovery/collector"
)

type Publisher interface {
	GetID() string
	GetInterval() time.Duration
	DiscoverAndPublish(ctx context.Context) (string, error)
}

type DefaultPublisher struct {
	id                string
	collectorClient   collector.Client
	discoveryFunction func(ctx context.Context) (any, string, error)
	interval          time.Duration
}

func NewDefaultPublisher(
	id string,
	interval time.Duration,
	collectorClient collector.Client,
	discoveryFunction func(ctx context.Context) (any, string, error),
) DefaultPublisher {
	return DefaultPublisher{
		id:                id,
		collectorClient:   collectorClient,
		discoveryFunction: discoveryFunction,
		interval:          interval,
	}
}

func (d DefaultPublisher) GetID() string {
	return d.id
}

func (d DefaultPublisher) GetInterval() time.Duration {
	return d.interval
}

func (d DefaultPublisher) DiscoverAndPublish(ctx context.Context) (string, error) {
	data, message, err := d.discoveryFunction(ctx)
	if err != nil {
		return "", err
	}

	err = d.collectorClient.Publish(ctx, d.id, data)
	if err != nil {
		return "", err
	}

	return message, nil
}
