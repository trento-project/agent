package factsengine

import (
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

const (
	factsGatheringRequestEvent = "trento.checks.v1.FactsRequest"
	factsGatheredEvent         = "trento.checks.v1.FactsGathered"
	eventSource                = "https://github.com/trento-project/agent/internal/factsengine"
)

func (c *FactsEngine) handleEvent(contentType string, request []byte) error {
	event, err := handleContentType(contentType, request)
	if err != nil {
		return errors.Wrap(err, "Error handling event")
	}

	switch eventType := event.Context.GetType(); eventType {
	case factsGatheringRequestEvent:
		err := c.handleFactsGatheringRequestEvent(event.DataEncoded)
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

func (c *FactsEngine) handleFactsGatheringRequestEvent(factsRequestByte []byte) error {
	var factsRequest gatherers.FactsRequest

	if err := ValidateAndUnmarshall(factsGatheringRequestEvent, factsRequestByte, &factsRequest); err != nil {
		return err
	}

	gatheredFacts, err := gatherFacts(c.agentID, factsRequest, c.factGatherers)
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

func buildFactsGatheredEvent(facts gatherers.FactsResult) ([]byte, error) {
	log.Infof("Building %s event...", factsGatheredEvent)

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetSource(eventSource)
	event.SetTime(time.Now())
	event.SetType(factsGatheredEvent)

	err := event.SetData(cloudevents.ApplicationJSON, facts)
	if err != nil {
		log.Fatalf("Failed to set data: %v", err)
	}

	log.Debugf("New event created:\n%s", event.String())

	jsonEvent, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	log.Infof("Event %s built properly", factsGatheredEvent)

	return jsonEvent, nil
}

func (c *FactsEngine) publishFacts(facts gatherers.FactsResult) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	response, err := buildFactsGatheredEvent(facts)
	if err != nil {
		log.Errorf("Error building response: %v", err)
		return err
	}

	if err := c.factsServiceAdapter.Publish(response, cloudevents.ApplicationCloudEventsJSON); err != nil {
		log.Error(err)
		return err
	}

	log.Infof("Gathered facts published properly")
	return nil
}
