package factsengine

import (
	"encoding/json"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/gatherers"

	contracts "github.com/trento-project/contracts/go/pkg/gen/entities"
)

func (c *FactsEngine) publishFacts(facts gatherers.FactsResult) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	data, err := json.Marshal(facts)
	if err != nil {
		return errors.Wrap(err, "could not serialize facts result")
	}

	event, err := contracts.NewFactsGatheredV1FromJson(data)
	if err != nil {
		log.Errorf("Error building event: %v", err)
		return err
	}

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
