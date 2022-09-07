package factsengine

import (
	"encoding/json"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"

	contracts "github.com/trento-project/contracts/go/pkg/gen/entities"
)

const (
	factsGatheringRequested = "trento.checks.v1.wanda.FactsGatheringRequested"
)

func (c *FactsEngine) handleEvent(contentType string, request []byte) error {
	event, err := handleContentType(contentType, request)
	if err != nil {
		return errors.Wrap(err, "Error handling event")
	}

	switch eventType := event.Context.GetType(); eventType {
	case factsGatheringRequested:
		err := c.handleFactsGatheringRequestedEvent(request)
		if err != nil {
			return errors.Wrap(err, "Error handling facts request")
		}
	default:
		return fmt.Errorf("Invalid event type: %s", eventType)
	}
	return nil
}

func handleContentType(contentType string, event []byte) (*cloudevents.Event, error) {
	switch contentType {
	case cloudevents.ApplicationCloudEventsJSON:
		event, err := handleCloudEvents(event)
		if err != nil {
			return nil, errors.Wrap(err, "Error handling cloudevent")
		}
		return &event, nil
	default:
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}
}

func handleCloudEvents(request []byte) (cloudevents.Event, error) {
	var event cloudevents.Event
	if err := json.Unmarshal(request, &event); err != nil {
		return event, errors.Wrap(err, "Error unmarshalling cloud event")
	}
	log.Debugf("New event received:\n%s", event.String())
	return event, nil
}

func (c *FactsEngine) handleFactsGatheringRequestedEvent(factsRequestByte []byte) error {
	event, err := contracts.NewFactsGatheringRequestedV1FromJsonCloudEvent(factsRequestByte)
	if err != nil {
		return err
	}

	factsRequest := entities.FactsGatheringRequestedFromEvent(event)
	agentFactsRequest := getAgentFacts(c.agentID, factsRequest)

	if agentFactsRequest == nil {
		log.Infof("FactsGatheringRequest is not for this agent. Discarding facts gathering execution")
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

func getAgentFacts(agentID string, factsRequest entities.FactsGatheringRequested) *entities.AgentFacts {
	for _, agentRequests := range factsRequest.Facts {
		if agentRequests.AgentID == agentID {
			return &agentRequests
		}
	}

	return nil
}

func (c *FactsEngine) publishFacts(facts entities.FactsGathered) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	event := entities.FactsGatheredToEvent(facts)

	serializedEvent, err := event.SerializeCloudEvent()
	if err != nil {
		log.Errorf("Error serializing event: %v", err)
		return err
	}

	if err := c.factsServiceAdapter.Publish(
		exchange, executionsRoutingKey, cloudevents.ApplicationCloudEventsJSON, serializedEvent); err != nil {

		log.Error(err)
		return err
	}

	log.Infof("Gathered facts published properly")
	return nil
}
