// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

// Package testsupport provides shared scaffolding for RabbitMQ-backed
// integration test suites (facts engine, operations, discovery).
package testsupport

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/v3/internal/messaging"
)

const defaultAMQPService = "amqp://guest:guest@localhost:5675" //nolint:gosec

// RabbitMQServiceURL returns the RABBITMQ_URL environment variable, falling
// back to the default used by the local docker-compose integration setup.
func RabbitMQServiceURL() string {
	amqpService := os.Getenv("RABBITMQ_URL")
	if amqpService == "" {
		amqpService = defaultAMQPService
	}

	return amqpService
}

// RabbitMQIntegrationSuite provides the setup/teardown boilerplate shared by
// integration test suites that listen for requests on a RabbitMQ queue.
// Embedding suites must set QueueName, ExchangeName and RoutingKey before
// suite.Run is called (typically in the Test<X>TestSuite entrypoint).
type RabbitMQIntegrationSuite struct {
	suite.Suite

	QueueName    string
	ExchangeName string
	RoutingKey   string

	AMQPService     string
	RabbitMQAdapter messaging.Adapter
}

func (s *RabbitMQIntegrationSuite) SetupSuite() {
	s.AMQPService = RabbitMQServiceURL()
}

func (s *RabbitMQIntegrationSuite) SetupTest() {
	rabbitmqAdapter, err := messaging.NewRabbitMQAdapter(
		s.AMQPService,
		s.QueueName,
		s.ExchangeName,
		s.RoutingKey,
	)
	if err != nil {
		panic(err)
	}

	s.RabbitMQAdapter = rabbitmqAdapter
}

func (s *RabbitMQIntegrationSuite) TearDownTest() {
	if s.RabbitMQAdapter == nil {
		return
	}

	err := s.RabbitMQAdapter.Unsubscribe()
	if err != nil {
		panic(err)
	}
}

// PublishUntilDone repeatedly publishes event on routingKey until ctx is
// done. The listener goroutine under test binds its queue asynchronously, so
// a single publish can race ahead of that binding and be silently dropped.
// Retrying until the request is picked up (or ctx expires) avoids relying on
// a fixed sleep. Publish errors are logged via warnMsg and retried rather
// than treated as fatal: transient reconnects on the underlying AMQP
// connection are expected and should not fail the test.
func PublishUntilDone(
	ctx context.Context,
	adapter messaging.Adapter,
	routingKey string,
	event []byte,
	warnMsg string,
) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for ctx.Err() == nil {
		err := adapter.Publish(routingKey, "", event)
		if err != nil {
			slog.Warn(warnMsg, "error", err)
		}

		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}
}
