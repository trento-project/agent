package discovery

import (
	"context"
	"fmt"
	"slices"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/messaging"

	"github.com/trento-project/contracts/go/pkg/events"
)

const (
	DiscoveryRequestedV1   string = "Trento.Discoveries.V1.DiscoveryRequested"
	exchange               string = "trento.discoveries"
	agentsQueue            string = "trento.discoveries.agents.%s"
	agentsEventsRoutingKey string = "agents"
)

func ListenRequests(
	ctx context.Context,
	agentID string,
	amqpServiceURL string,
	discoveries []Discovery,
) error {
	log.Infof("Subscribing agent %s to the discovery requests on %s", agentID, amqpServiceURL)
	queue := fmt.Sprintf(agentsQueue, agentID)
	amqpAdapter, err := messaging.NewRabbitMQAdapter(
		amqpServiceURL,
		queue,
		exchange,
		agentsEventsRoutingKey,
	)
	if err != nil {
		return err
	}

	log.Infof("Listening for discovery requests...")
	defer func() {
		if err = amqpAdapter.Unsubscribe(); err != nil {
			log.Errorf("Error during unsubscription: %s", err)
		}
	}()

	discoveriesMap := make(map[string]Discovery)
	for _, d := range discoveries {
		discoveriesMap[d.GetID()] = d
	}

	if err := amqpAdapter.Listen(
		func(_ string, event []byte) error {
			return HandleEvent(ctx, event, agentID, discoveriesMap)
		}); err != nil {
		return err
	}

	<-ctx.Done()

	return err
}

func HandleEvent(
	ctx context.Context,
	event []byte,
	agentID string,
	discoveries map[string]Discovery,
) error {
	log.Info("New DiscoveryRequested message received")
	eventType, err := events.EventType(event)
	if err != nil {
		return errors.Wrap(err, "error getting event type")
	}
	switch eventType {
	case DiscoveryRequestedV1:
		discoveryRequested, err := DiscoveryRequestedFromEvent(event)
		if err != nil {
			return errors.Wrap(err, "error decoding DiscoveryRequested event")
		}

		if !slices.Contains(discoveryRequested.Targets, agentID) {
			log.Infof("DiscoveryRequested is not for this agent. Discarding request")
			return nil
		}

		requestedDiscovery, found := discoveries[discoveryRequested.DiscoveryType]
		if !found {
			return fmt.Errorf("unknown discovery type: %s", discoveryRequested.DiscoveryType)
		}

		// Run discovery
		message, err := requestedDiscovery.Discover(ctx)
		if err != nil {
			return errors.Wrap(err, "error during discovery")

		}
		log.Info(message)
		return nil
	default:
		return fmt.Errorf("invalid event type: %s", eventType)
	}
}
