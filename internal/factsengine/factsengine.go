package factsengine

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/wagslane/go-rabbitmq"
)

const (
	gatherFactsExchanage string = "gather_facts"
	factsExchanage       string = "facts"
)

type factsEngine struct {
	agentID            string
	factsEngineService string
	factGatherers      map[string]gatherers.FactGatherer
	consumer           rabbitmq.Consumer
	publisher          *rabbitmq.Publisher
}

func NewFactsEngine(agentID, factsEngineService string) *factsEngine {
	return &factsEngine{
		agentID:            agentID,
		factsEngineService: factsEngineService,
		factGatherers: map[string]gatherers.FactGatherer{
			gatherers.CorosyncFactKey: gatherers.NewCorosyncConfGatherer(),
		},
	}
}

func (c *factsEngine) Subscribe() error {
	log.Infof("Subscribing agent %s to the facts gathering reception service on %s", c.agentID, c.factsEngineService)
	consumer, err := rabbitmq.NewConsumer(
		c.factsEngineService,
		rabbitmq.Config{},
		rabbitmq.WithConsumerOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	c.consumer = consumer

	publisher, err := rabbitmq.NewPublisher(
		c.factsEngineService,
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	c.publisher = publisher

	log.Infof("Subscription to the facts engine by agent %s in %s done", c.agentID, c.factsEngineService)

	return nil
}

func (c *factsEngine) Unsubscribe() error {
	log.Infof("Unsubscribing agent %s from the facts engine service", c.agentID)
	c.consumer.Close()
	c.publisher.Close()
	log.Infof("Unsubscribed properly")

	return nil
}

func (c *factsEngine) Listen(ctx context.Context) {
	log.Infof("Listening for facts gathering events...")
	defer c.Unsubscribe()

	err := c.consumer.StartConsuming(
		func(d rabbitmq.Delivery) rabbitmq.Action {
			factsRequests, err := parseFactsRequest(d.Body)
			if err != nil {
				log.Errorf("Invalid facts request: %s", err)
				return rabbitmq.NackDiscard
			}

			gatheredFacts, _ := gatherFacts(factsRequests, c.factGatherers)
			c.publishFacts(gatheredFacts)
			return rabbitmq.Ack
		},
		c.agentID,
		[]string{c.agentID},
		rabbitmq.WithConsumeOptionsQueueDurable,
		rabbitmq.WithConsumeOptionsQueueAutoDelete,
		rabbitmq.WithConsumeOptionsBindingExchangeName(gatherFactsExchanage),
		rabbitmq.WithConsumeOptionsBindingExchangeDurable,
		rabbitmq.WithConsumeOptionsBindingExchangeAutoDelete,
	)
	if err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}

func gatherFacts(groupedFactsRequest *gatherers.GroupedFactsRequest, factGatherers map[string]gatherers.FactGatherer) (*gatherers.FactsResult, error) {
	factsResults := &gatherers.FactsResult{
		ExecutionID: groupedFactsRequest.ExecutionID,
	}
	log.Infof("Starting facts gathering process")

	// Gather facts asynchronously
	var wg sync.WaitGroup

	for gathererType, factsRequest := range groupedFactsRequest.Facts {

		g, exists := factGatherers[gathererType]
		if !exists {
			log.Errorf("Fact gatherer %s does not exist", gathererType)
			continue
		}

		// Execute the fact gathering asynchronously and in parallel
		wg.Add(1)
		go func(wg *sync.WaitGroup, factRequest []*gatherers.FactRequest) {
			defer wg.Done()
			newFacts, err := g.Gather(factRequest)
			if err == nil {
				factsResults.Facts = append(factsResults.Facts, newFacts[:]...)
			} else {
				log.Error(err)
			}
		}(&wg, factsRequest)
	}

	wg.Wait()

	log.Infof("Requested facts gathered")
	return factsResults, nil
}

func parseFactsRequest(request []byte) (*gatherers.GroupedFactsRequest, error) {
	var factsRequest gatherers.FactsRequest
	var groupedFactsRequest *gatherers.GroupedFactsRequest

	err := json.Unmarshal(request, &factsRequest)
	if err != nil {
		return nil, err
	}

	groupedFactsRequest = &gatherers.GroupedFactsRequest{
		ExecutionID: factsRequest.ExecutionID,
		Facts:       make(map[string][]*gatherers.FactRequest),
	}

	// Group the received facts by gatherer type, so they are executed in the same moment with the same source of truth
	for _, factRequest := range factsRequest.Facts {
		groupedFactsRequest.Facts[factRequest.Gatherer] = append(groupedFactsRequest.Facts[factRequest.Gatherer], factRequest)
	}

	return groupedFactsRequest, nil
}

func buildResponse(facts *gatherers.FactsResult) ([]byte, error) {
	log.Infof("Building gathered facts response...")

	jsonFacts, err := json.Marshal(facts)
	if err != nil {
		return nil, err
	}

	log.Infof("Gathered facts response built properly")

	return jsonFacts, nil
}

func prettyString(str []byte) string {
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, str, "", "  ")
	return prettyJSON.String()
}

func (c *factsEngine) publishFacts(facts *gatherers.FactsResult) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	response, err := buildResponse(facts)
	if err != nil {
		log.Errorf("Error building response: %v", err)
		return err
	}

	log.Debugf("Gathered facts response: %s", prettyString(response))
	err = c.publisher.Publish(
		response,
		[]string{""},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsExchange(factsExchanage),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Infof("Gathered facts published properly")
	return nil
}
