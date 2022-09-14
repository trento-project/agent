package factsengine

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
	"github.com/trento-project/contracts/golang/pkg/events"
)

func (c *FactsEngine) handleEvent(contentType string, request []byte) error {
	eventType, err := events.EventType(request)
	if err != nil {
		return errors.Wrap(err, "Error getting event type")
	}
	switch eventType {
	case events.EventTypeFromProto(&events.FactsGatheringRequested{}):
		err := c.handleFactsGatheringRequestedEvent(request)
		if err != nil {
			return errors.Wrap(err, "Error handling facts request")
		}
	default:
		return fmt.Errorf("Invalid event type: %s", eventType)
	}
	return nil
}

func (c *FactsEngine) handleFactsGatheringRequestedEvent(factsRequestByte []byte) error {
	factsRequest, err := entities.FactsGatheringRequestedFromEvent(factsRequestByte)
	if err != nil {
		return err
	}

	agentFactsRequest := getAgentFacts(c.agentID, factsRequest)

	if agentFactsRequest == nil {
		log.Infof("FactsGatheringRequested is not for this agent. Discarding facts gathering execution")
		return nil
	}

	gatheredFacts, err := gatherFacts(factsRequest.ExecutionID, c.agentID, agentFactsRequest, c.factGatherers)
	if err != nil {
		log.Errorf("Error gathering facts: %s", err)
		return err
	}

	if err := c.publishFacts(gatheredFacts); err != nil {
		log.Errorf("Error publishing facts: %s", err)
		return err
	}

	return nil
}

func getAgentFacts(agentID string, factsRequest *entities.FactsGatheringRequested) *entities.FactsGatheringRequestedTarget {
	for _, agentRequests := range factsRequest.Targets {
		if agentRequests.AgentID == agentID {
			return &agentRequests
		}
	}

	return nil
}

func (c *FactsEngine) publishFacts(facts entities.FactsGathered) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	event, err := entities.FactsGatheredToEvent(facts)
	if err != nil {
		return err
	}

	if err := c.factsServiceAdapter.Publish(
		exchange, executionsRoutingKey, cloudevents.ApplicationCloudEventsJSON, event); err != nil {

		log.Error(err)
		return err
	}

	log.Infof("Gathered facts published properly")
	return nil
}
