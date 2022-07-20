package factsengine

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

type FactsEngineTestSuite struct {
	suite.Suite
}

func TestFactsEngineTestSuite(t *testing.T) {
	suite.Run(t, new(FactsEngineTestSuite))
}

type dummyGatherer1 struct {
}

func NewDummyGatherer1() *dummyGatherer1 {
	return &dummyGatherer1{}
}

func (s *dummyGatherer1) Gather(_ []gatherers.FactRequest) ([]gatherers.Fact, error) {
	return []gatherers.Fact{
		{
			Name:  "dummy1",
			Value: "1",
		},
	}, nil
}

type dummyGatherer2 struct {
}

func NewDummyGatherer2() *dummyGatherer2 {
	return &dummyGatherer2{}
}

func (s *dummyGatherer2) Gather(_ []gatherers.FactRequest) ([]gatherers.Fact, error) {
	return []gatherers.Fact{
		{
			Name:  "dummy2",
			Value: "2",
		},
	}, nil
}

type errorGatherer struct {
}

func NewErrorGatherer() *errorGatherer {
	return &errorGatherer{}
}

func (s *errorGatherer) Gather(_ []gatherers.FactRequest) ([]gatherers.Fact, error) {
	return []gatherers.Fact{}, fmt.Errorf("kabum!")
}

func (suite *FactsEngineTestSuite) TestCorosyncConf_GatherFacts() {
	someID := "someID"

	groupedFactsRequest := &gatherers.GroupedFactsRequest{
		ExecutionID: someID,
		Facts: map[string][]gatherers.FactRequest{
			"dummyGatherer1": {
				{
					Name:     "dummy1",
					Gatherer: "dummyGatherer1",
					Argument: "dummy1",
				},
			},
			"dummyGatherer2": {
				{
					Name:     "dummy2",
					Gatherer: "dummyGatherer2",
					Argument: "dummy2",
				},
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"dummyGatherer2": NewDummyGatherer2(),
	}

	factResults, err := gatherFacts(groupedFactsRequest, factGatherers)

	expectedFacts := []gatherers.Fact{
		{
			Name:  "dummy1",
			Value: "1",
		},
		{
			Name:  "dummy2",
			Value: "2",
		},
	}

	suite.NoError(err)
	suite.Equal(someID, factResults.ExecutionID)
	suite.ElementsMatch(expectedFacts, factResults.Facts)
}

func (suite *FactsEngineTestSuite) TestCorosyncConf_GatherFacts_GathererNotFound() {
	someID := "someID"

	groupedFactsRequest := &gatherers.GroupedFactsRequest{
		ExecutionID: someID,
		Facts: map[string][]gatherers.FactRequest{
			"dummyGatherer1": {
				{
					Name:     "dummy1",
					Gatherer: "dummyGatherer1",
					Argument: "dummy1",
				},
			},
			"otherGatherer": {
				{
					Name:     "other",
					Gatherer: "otherGatherer",
					Argument: "other",
				},
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"dummyGatherer2": NewDummyGatherer2(),
	}

	factResults, err := gatherFacts(groupedFactsRequest, factGatherers)

	expectedFacts := []gatherers.Fact{
		{
			Name:  "dummy1",
			Value: "1",
		},
	}

	suite.NoError(err)
	suite.Equal(someID, factResults.ExecutionID)
	suite.ElementsMatch(expectedFacts, factResults.Facts)
}

func (suite *FactsEngineTestSuite) TestCorosyncConf_GatherFacts_ErrorGathering() {
	someID := "someID"

	groupedFactsRequest := &gatherers.GroupedFactsRequest{
		ExecutionID: someID,
		Facts: map[string][]gatherers.FactRequest{
			"dummyGatherer1": {
				{
					Name:     "dummy1",
					Gatherer: "dummyGatherer1",
					Argument: "dummy1",
				},
			},
			"errorGatherer": {
				{
					Name:     "error",
					Gatherer: "errorGatherer",
					Argument: "error",
				},
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"errorGatherer":  NewErrorGatherer(),
	}

	factResults, err := gatherFacts(groupedFactsRequest, factGatherers)

	expectedFacts := []gatherers.Fact{
		{
			Name:  "dummy1",
			Value: "1",
		},
	}

	suite.NoError(err)
	suite.Equal(someID, factResults.ExecutionID)
	suite.ElementsMatch(expectedFacts, factResults.Facts)
}

func (suite *FactsEngineTestSuite) TestCorosyncConf_ParseFactsRequest() {

	factsRequests := `
	{
		"execution_id": "some-id",
		"facts": [
			{"name": "sbd_device", "gatherer": "sbd_config", "argument": "SBD_DEVICE"},
			{"name": "sbd_timeout_actions", "gatherer": "sbd_config", "argument": "SBD_TIMEOUT_ACTION"},
			{"name": "pacemaker_version", "gatherer": "package_version", "argument": "pacemaker"},
			{"name": "corosync_version", "gatherer": "package_version", "argument": "corosync"},
			{"name": "sbd_pcmk_delay_max", "gatherer": "cib", "argument": "//primitive[@type='external/sbd']/instance_attributes/nvpair[@name='pcmk_delay_max']/@value"},
			{"name": "cib_sid", "gatherer": "cib", "argument": "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='SID']/@value"},
			{"name": "corosync_token", "gatherer": "corosync.conf", "argument": "totem.token"},
			{"name": "corosync_join", "gatherer": "corosync.conf", "argument": "totem.join"}
		]
	}`

	groupedFactRequsets, err := parseFactsRequest([]byte(factsRequests))

	expectedRequests := &gatherers.GroupedFactsRequest{
		ExecutionID: "some-id",
		Facts: map[string][]gatherers.FactRequest{
			"sbd_config": {
				{
					Name:     "sbd_device",
					Gatherer: "sbd_config",
					Argument: "SBD_DEVICE",
				},
				{
					Name:     "sbd_timeout_actions",
					Gatherer: "sbd_config",
					Argument: "SBD_TIMEOUT_ACTION",
				},
			},
			"package_version": {
				{
					Name:     "pacemaker_version",
					Gatherer: "package_version",
					Argument: "pacemaker",
				},
				{
					Name:     "corosync_version",
					Gatherer: "package_version",
					Argument: "corosync",
				},
			},
			"cib": {
				{
					Name:     "sbd_pcmk_delay_max",
					Gatherer: "cib",
					Argument: "//primitive[@type='external/sbd']/instance_attributes/nvpair[@name='pcmk_delay_max']/@value",
				},
				{
					Name:     "cib_sid",
					Gatherer: "cib",
					Argument: "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='SID']/@value",
				},
			},
			"corosync.conf": {
				{
					Name:     "corosync_token",
					Gatherer: "corosync.conf",
					Argument: "totem.token",
				},
				{
					Name:     "corosync_join",
					Gatherer: "corosync.conf",
					Argument: "totem.join",
				},
			},
		},
	}

	suite.NoError(err)
	suite.Equal(expectedRequests, groupedFactRequsets)
}

func (suite *FactsEngineTestSuite) TestCorosyncConf_BuildResponse() {
	facts := gatherers.FactsResult{
		ExecutionID: "some-id",
		Facts: []gatherers.Fact{
			{
				Name:  "fact1",
				Value: "1",
			},
			{
				Name:  "fact2",
				Value: "2",
			},
		},
	}

	response, err := buildResponse(facts)

	expectedResponse := `{"execution_id":"some-id","facts":[{"name":"fact1","value":"1"},{"name":"fact2","value":"2"}]}`

	suite.NoError(err)
	suite.Equal(expectedResponse, string(response))
}
