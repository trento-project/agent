package checksengine

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"

	"github.com/trento-project/agent/internal/checksengine/facts"
)

const fakeRequest string = `[
{"name": "sbd_device", "gatherer": "sbd_config", "argument": "SBD_DEVICE"},
{"name": "sbd_timeout_actions", "gatherer": "sbd_config", "argument": "SBD_TIMEOUT_ACTION"},
{"name": "pacemaker_version", "gatherer": "package_version", "argument": "pacemaker"},
{"name": "corosync_version", "gatherer": "package_version", "argument": "corosync"},
{"name": "other_version", "gatherer": "package_version", "argument": "other"},
{"name": "sbd_pcmk_delay_max", "gatherer": "cib", "argument": "//primitive[@type='external/sbd']/instance_attributes/nvpair[@name='pcmk_delay_max']/@value"},
{"name": "cib_sid", "gatherer": "cib", "argument": "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='SID']/@value"},
{"name": "cib_instance_number", "gatherer": "cib", "argument": "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='InstanceNumber']/@value"},
{"name": "cib_saphana_start_interval", "gatherer": "cib", "argument": "//primitive[@type='SAPHana']/operations/op[@name='start']/@interval"},
{"name": "cib_saphana_start_timeout", "gatherer": "cib", "argument": "//primitive[@type='SAPHana']/operations/op[@name='start']/@timeout"},
{"name": "cib_saphana_monitor_master_timeout", "gatherer": "cib", "argument": "//primitive[@type='SAPHana']/operations/op[@name='monitor' and @role='Master']/@timeout"},
{"name": "crmmon_sbd_role", "gatherer": "crmmon", "argument": "//resource[@resource_agent='stonith:external/sbd']/@role"},
{"name": "corosync_token", "gatherer": "corosync.conf", "argument": "totem.token"},
{"name": "corosync_join", "gatherer": "corosync.conf", "argument": "totem.join"},
{"name": "corosync_node1id", "gatherer": "corosync.conf", "argument": "nodelist.node.0.nodeid"},
{"name": "corosync_node2id", "gatherer": "corosync.conf", "argument": "nodelist.node.1.nodeid"},
{"name": "corosync_nodes", "gatherer": "corosync.conf", "argument": "nodelist.node"},
{"name": "corosync_not_found", "gatherer": "corosync.conf", "argument": "totem.not_found"}
]`

const gatherFactsExchanage string = "gather_facts"
const factsExchanage string = "facts"

type checksEngine struct {
	agentID             string
	checksEngineService string
	gatherers           map[string]facts.FactGatherer
	consumer            rabbitmq.Consumer
	publisher           *rabbitmq.Publisher
}

func NewChecksEngine(agentID, checksEngineService string) *checksEngine {
	return &checksEngine{
		agentID:             agentID,
		checksEngineService: checksEngineService,
		gatherers: map[string]facts.FactGatherer{
			facts.SBDFactKey:            facts.NewSbdConfigGatherer(),
			facts.PackageVersionFactKey: facts.NewPackageVersionConfigGatherer(),
			facts.CibFactKey:            facts.NewCibConfigGatherer(),
			facts.CrmmonFactKey:         facts.NewCrmmonConfigGatherer(),
			facts.CorosyncFactKey:       facts.NewCorosyncConfGatherer(),
		},
	}
}

func (c *checksEngine) Subscribe() error {
	log.Infof("Subscribing agent %s to the checks engine runner on %s", c.agentID, c.checksEngineService)
	consumer, err := rabbitmq.NewConsumer(
		c.checksEngineService,
		rabbitmq.Config{},
		rabbitmq.WithConsumerOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	c.consumer = consumer

	publisher, err := rabbitmq.NewPublisher(
		c.checksEngineService,
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	c.publisher = publisher

	log.Infof("Subscription to the checks engine by agent %s in %s done", c.agentID, c.checksEngineService)

	return nil
}

func (c *checksEngine) Unsubscribe() error {
	log.Infof("Unsubscribing agent %s to the checks engine runner", c.agentID)
	c.consumer.Close()
	c.publisher.Close()
	log.Infof("Unsubscribed properly")

	return nil

}

func (c *checksEngine) Listen(ctx context.Context) {
	log.Infof("Listening for checks execution events...")
	defer c.Unsubscribe()

	// Dummy code to gather SBD configuration files every some seconds
	//c.dummyGatherer(ctx)

	err := c.consumer.StartConsuming(
		func(d rabbitmq.Delivery) rabbitmq.Action {
			factsRequests, err := parseFactsRequest(d.Body)
			if err != nil {
				log.Errorf("Invalid facts request: %s", err)
				return rabbitmq.NackDiscard
			}

			gatheredFacts, _ := gatherFacts(factsRequests, c.gatherers)
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

func (c *checksEngine) dummyGatherer(ctx context.Context) {
	factsRequests, err := parseFactsRequest([]byte(fakeRequest))
	if err != nil {
		log.Errorf("Invalid facts request: %s", err)
		return
	}

	ticker := time.NewTicker(10 * time.Second)

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			gatheredFacts, _ := gatherFacts(factsRequests, c.gatherers)
			c.publishFacts(gatheredFacts)
		case <-ctx.Done():
			return
		}
	}
}

func gatherFacts(groupedFactsRequest facts.GroupedFactsRequest, gatherers map[string]facts.FactGatherer) ([]*facts.Fact, error) {
	var gatheredFacts []*facts.Fact
	log.Infof("Starting facts gathering process")

	// Gather facts asynchronously
	var wg sync.WaitGroup

	for gathererType, factsRequest := range groupedFactsRequest {

		g, exists := gatherers[gathererType]
		if !exists {
			log.Errorf("Fact gatherer %s does not exist", gathererType)
			continue
		}

		// Execute the fact gathering asynchronously and in parallel
		wg.Add(1)
		go func(wg *sync.WaitGroup, factRequest []*facts.FactRequest) {
			defer wg.Done()
			newFacts, err := g.Gather(factRequest)
			if err == nil {
				gatheredFacts = append(gatheredFacts, newFacts[:]...)
			} else {
				log.Error(err)
			}
		}(&wg, factsRequest)
	}

	wg.Wait()

	log.Infof("Requested facts gathered")
	return gatheredFacts, nil
}

func parseFactsRequest(request []byte) (facts.GroupedFactsRequest, error) {
	var factsRequest []*facts.FactRequest
	groupedFactsRequest := make(facts.GroupedFactsRequest)

	err := json.Unmarshal(request, &factsRequest)
	if err != nil {
		return nil, err
	}

	// Group the received facts by gatherer type, so they are executed in the same moment with the same source of truth
	for _, factRequest := range factsRequest {
		groupedFactsRequest[factRequest.Gatherer] = append(groupedFactsRequest[factRequest.Gatherer], factRequest)
	}

	return groupedFactsRequest, nil
}

func buildResponse(facts []*facts.Fact) ([]byte, error) {
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

func (c *checksEngine) publishFacts(facts []*facts.Fact) error {
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

	// Publish somehow the gathered facts
	log.Infof("Gathered facts published properly")
	return nil
}
