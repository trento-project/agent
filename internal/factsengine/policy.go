package factsengine

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

func (c *FactsEngine) publishFacts(facts gatherers.FactsResult) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	event := gatherers.FactsGatheredToEvent(facts)

	serializedEvent, err := event.SerializeCloudEvent()
	if err != nil {
		log.Errorf("Error serializing event: %v", err)
		return err
	}

	if err := c.factsServiceAdapter.Publish(cloudevents.ApplicationCloudEventsJSON, serializedEvent); err != nil {
		log.Error(err)
		return err
	}

	log.Infof("Gathered facts published properly")
	return nil
}
