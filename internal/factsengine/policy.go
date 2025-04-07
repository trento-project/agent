package factsengine

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/messaging"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/contracts/go/pkg/events"
)

const (
	FactsGatheringRequested = "Trento.Checks.V1.FactsGatheringRequested"
)

func HandleEvent(
	ctx context.Context,
	event []byte,
	agentID string,
	adapter messaging.Adapter,
	registry gatherers.Registry,
) error {
	eventType, err := events.EventType(event)
	if err != nil {
		return errors.Wrap(err, "Error getting event type")
	}
	switch eventType {
	case FactsGatheringRequested:
		factsRequest, err := FactsGatheringRequestedFromEvent(event)
		if err != nil {
			return err
		}

		agentFactsRequest := getAgentFacts(agentID, factsRequest)

		if agentFactsRequest == nil {
			log.Infof("FactsGatheringRequested is not for this agent. Discarding facts gathering execution")
			return nil
		}

		gatheredFacts, err := gatherFacts(
			ctx,
			factsRequest.ExecutionID,
			agentID,
			factsRequest.GroupID,
			agentFactsRequest,
			registry,
		)
		if err != nil {
			log.Errorf("Error gathering facts: %s", err)
			return errors.Wrap(err, "Error gathering facts")
		}

		log.Infof("Publishing gathered facts to the checks engine service")
		event, err := FactsGatheredToEvent(gatheredFacts)
		if err != nil {
			return errors.Wrap(err, "Error encoding gathered facts")
		}

		if err := adapter.Publish(executionsRoutingKey, events.ContentType(), event); err != nil {
			log.Error(err)
			return errors.Wrap(err, "Error publishing gathered facts")
		}

		log.Infof("Gathered facts published properly")

		return nil
	default:
		return fmt.Errorf("Invalid event type: %s", eventType)
	}
}

func getAgentFacts(
	agentID string,
	factsRequest *entities.FactsGatheringRequested) *entities.FactsGatheringRequestedTarget {

	for _, agentRequests := range factsRequest.Targets {
		if agentRequests.AgentID == agentID {
			return &agentRequests
		}
	}

	return nil
}
