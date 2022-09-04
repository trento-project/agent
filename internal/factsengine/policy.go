package factsengine

import (
	"encoding/base64"
	"encoding/json"

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

	var deserializedEvent map[string]interface{}
	var data map[string]interface{}

	err = json.Unmarshal(serializedEvent, &deserializedEvent)
	if err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(deserializedEvent["data_base64"].(string))

	err = json.Unmarshal(decoded, &data)
	if err != nil {
		return err
	}

	deserializedEvent["data"] = data

	delete(deserializedEvent, "data_base64")

	serializedEvent, _ = json.Marshal(deserializedEvent)

	if err := c.factsServiceAdapter.Publish(
		factsExchange, cloudevents.ApplicationCloudEventsJSON, serializedEvent); err != nil {

		log.Error(err)
		return err
	}

	log.Infof("Gathered facts published properly")
	return nil
}
