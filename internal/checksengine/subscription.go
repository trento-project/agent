package checksengine

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/trento-project/agent/internal/checksengine/facts"
)

type checksEngine struct {
	agentID             string
	checksEngineService string
	gatherers           map[string]facts.FactGatherer
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
	// Subscribe somehow to the checks engine runner
	log.Infof("Subscription to the checks engine by agent %s in %s done", c.agentID, c.checksEngineService)

	return nil
}

func (c *checksEngine) Unsubscribe() error {
	log.Infof("Unsubscribing agent %s to the checks engine runner", c.agentID)
	// Unsubscribe somehow from the checks engine runner
	log.Infof("Unsubscribed properly")

	return nil

}

func (c *checksEngine) Listen(ctx context.Context) {
	log.Infof("Listening for checks execution events...")
	defer c.Unsubscribe()

	// Dummy code to gather SBD configuration files every some seconds
	c.dummyGatherer(ctx)
}

func (c *checksEngine) dummyGatherer(ctx context.Context) {
	rawFactsRequests :=
		`[
{"type": "sbd_config", "facts": [{"name":"SBD_DEVICE"}, {"name":"SBD_TIMEOUT_ACTION"}]},
{"type": "package_version", "facts": [{"name":"pacemaker"}, {"name":"corosync"}, {"name":"other"}]},
{"type": "cib", "facts": [
	{"name":"//primitive[@type='external/sbd']/instance_attributes/nvpair[@name='pcmk_delay_max']/@value", "alias": "sbd_pcmk_delay_max"},
	{"name":"//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='SID']/@value", "alias": "cib_sid"},
	{"name":"//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='InstanceNumber']/@value", "alias": "cib_instance_number"},
	{"name":"//primitive[@type='SAPHana']/operations/op[@name='start']/@interval", "alias": "cib_saphana_start_interval"},
	{"name":"//primitive[@type='SAPHana']/operations/op[@name='start']/@timeout", "alias": "cib_saphana_start_timeout"},
	{"name":"//primitive[@type='SAPHana']/operations/op[@name='monitor' and @role='Master']/@timeout", "alias": "cib_saphana_monitor_master_timeout"}
]},
{"type": "crmmon", "facts": [
	{"name":"//resource[@resource_agent='stonith:external/sbd']/@role", "alias": "crmmon_sbd_role"}
]},
{"type": "corosync.conf", "facts": [
	{"name":"totem.token", "alias": "corosync_token"},
	{"name":"totem.join", "alias": "corosync_join"},
	{"name":"nodelist.node.0.nodeid", "alias": "corosync_node1id"},
	{"name":"nodelist.node.1.nodeid", "alias": "corosync_node2id"},
	{"name":"nodelist.node", "alias": "corosync_nodes"},
	{"name":"totem.not_found", "alias": "corosync_not_found"}
]}]`

	factsRequests, err := parseFactsRequest([]byte(rawFactsRequests))
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
			publishFacts(gatheredFacts)
		case <-ctx.Done():
			return
		}
	}
}

func gatherFacts(factsRequests []*facts.FactsRequest, gatherers map[string]facts.FactGatherer) ([]*facts.Fact, error) {
	var gatheredFacts []*facts.Fact
	log.Infof("Starting facts gathering process")

	// Gather facts asynchronously
	var wg sync.WaitGroup

	for _, factRequest := range factsRequests {

		g, exists := gatherers[factRequest.Type]
		if !exists {
			log.Errorf("Fact gatherer %s does not exist", factRequest.Type)
			continue
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, factRequest []facts.FactRequest) {
			defer wg.Done()
			newFacts, err := g.Gather(factRequest)
			if err == nil {
				gatheredFacts = append(gatheredFacts, newFacts[:]...)
			} else {
				log.Error(err)
			}
		}(&wg, factRequest.Facts)
	}

	wg.Wait()

	log.Infof("Requested facts gathered")
	return gatheredFacts, nil
}

func parseFactsRequest(request []byte) ([]*facts.FactsRequest, error) {
	var factsRequest []*facts.FactsRequest
	err := json.Unmarshal(request, &factsRequest)
	if err != nil {
		return nil, err
	}
	return factsRequest, nil
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

func publishFacts(facts []*facts.Fact) error {
	log.Infof("Publishing gathered facts to the checks engine service")
	response, err := buildResponse(facts)
	if err != nil {
		log.Errorf("Error building response: %v", err)
		return err
	}

	// By now, simply print the gathered facts
	log.Infof("Gathered facts response: %s", response)

	// Publish somehow the gathered facts
	log.Infof("Gathered facts published properly")
	return nil
}
