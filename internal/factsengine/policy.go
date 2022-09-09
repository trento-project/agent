package factsengine

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

func (c *FactsEngine) publishFacts(facts entities.FactsGathered) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	event := entities.FactsGatheredToEvent(facts)

	serializedEvent, err := event.SerializeCloudEvent()
	if err != nil {
		log.Errorf("Error serializing event: %v", err)
		return err
	}

	if err := c.factsServiceAdapter.Publish(
		factsExchange, cloudevents.ApplicationCloudEventsJSON, serializedEvent); err != nil {

		log.Error(err)
		return err
	}

	log.Infof("Gathered facts published properly")
	return nil
}
