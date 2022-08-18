//nolint:lll
package factsengine

import (
	"fmt"
	"os"
	"path"
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

type DummyGatherer1 struct {
}

func NewDummyGatherer1() *DummyGatherer1 {
	return &DummyGatherer1{}
}

func (s *DummyGatherer1) Gather(_ []gatherers.FactRequest) ([]gatherers.Fact, error) {
	return []gatherers.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
	}, nil
}

type DummyGatherer2 struct {
}

func NewDummyGatherer2() *DummyGatherer2 {
	return &DummyGatherer2{}
}

func (s *DummyGatherer2) Gather(_ []gatherers.FactRequest) ([]gatherers.Fact, error) {
	return []gatherers.Fact{
		{
			Name:    "dummy2",
			Value:   "2",
			CheckID: "check1",
		},
	}, nil
}

type ErrorGatherer struct {
}

func NewErrorGatherer() *ErrorGatherer {
	return &ErrorGatherer{}
}

func (s *ErrorGatherer) Gather(_ []gatherers.FactRequest) ([]gatherers.Fact, error) {
	return []gatherers.Fact{}, fmt.Errorf("kabum!") //nolint
}

func (suite *FactsEngineTestSuite) TestFactsEngineGatherFacts() {
	someID := "someID"     //nolint
	agentID := "someAgent" //nolint

	groupedFactsRequest := &gatherers.GroupedFactsRequest{
		ExecutionID: someID,
		Facts: map[string][]gatherers.FactRequest{
			"dummyGatherer1": {
				{
					Name:     "dummy1",
					Gatherer: "dummyGatherer1",
					Argument: "dummy1",
					CheckID:  "check1",
				},
			},
			"dummyGatherer2": {
				{
					Name:     "dummy2",
					Gatherer: "dummyGatherer2",
					Argument: "dummy2",
					CheckID:  "check1",
				},
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"dummyGatherer2": NewDummyGatherer2(),
	}

	factResults, err := gatherFacts(agentID, groupedFactsRequest, factGatherers)

	expectedFacts := []gatherers.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
		{
			Name:    "dummy2",
			Value:   "2",
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(someID, factResults.ExecutionID)
	suite.Equal(agentID, factResults.AgentID)
	suite.ElementsMatch(expectedFacts, factResults.Facts)
}

func (suite *FactsEngineTestSuite) TestFactsEngineGatherFactsGathererNotFound() {
	someID := "someID"
	agentID := "someAgent"

	groupedFactsRequest := &gatherers.GroupedFactsRequest{
		ExecutionID: someID,
		Facts: map[string][]gatherers.FactRequest{
			"dummyGatherer1": {
				{
					Name:     "dummy1",
					Gatherer: "dummyGatherer1",
					Argument: "dummy1",
					CheckID:  "check1",
				},
			},
			"otherGatherer": {
				{
					Name:     "other",
					Gatherer: "otherGatherer",
					Argument: "other",
					CheckID:  "check1",
				},
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"dummyGatherer2": NewDummyGatherer2(),
	}

	factResults, err := gatherFacts(agentID, groupedFactsRequest, factGatherers)

	expectedFacts := []gatherers.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(someID, factResults.ExecutionID)
	suite.Equal(agentID, factResults.AgentID)
	suite.ElementsMatch(expectedFacts, factResults.Facts)
}

func (suite *FactsEngineTestSuite) TestFactsEngineGatherFactsErrorGathering() {
	someID := "someID"
	agentID := "someAgent"

	groupedFactsRequest := &gatherers.GroupedFactsRequest{
		ExecutionID: someID,
		Facts: map[string][]gatherers.FactRequest{
			"dummyGatherer1": {
				{
					Name:     "dummy1",
					Gatherer: "dummyGatherer1",
					Argument: "dummy1",
					CheckID:  "check1",
				},
			},
			"errorGatherer": {
				{
					Name:     "error",
					Gatherer: "errorGatherer",
					Argument: "error",
					CheckID:  "check1",
				},
			},
		},
	}

	factGatherers := map[string]gatherers.FactGatherer{
		"dummyGatherer1": NewDummyGatherer1(),
		"errorGatherer":  NewErrorGatherer(),
	}

	factResults, err := gatherFacts(agentID, groupedFactsRequest, factGatherers)

	expectedFacts := []gatherers.Fact{
		{
			Name:    "dummy1",
			Value:   "1",
			CheckID: "check1",
		},
	}

	suite.NoError(err)
	suite.Equal(someID, factResults.ExecutionID)
	suite.Equal(agentID, factResults.AgentID)
	suite.ElementsMatch(expectedFacts, factResults.Facts)
}

func (suite *FactsEngineTestSuite) TestFactsEngineParseFactsRequest() {

	factsRequests := `
	{
		"execution_id": "some-id",
		"facts": [
			{"name": "sbd_device", "gatherer": "sbd_config", "argument": "SBD_DEVICE", "check_id": "check1"},
			{"name": "sbd_timeout_actions", "gatherer": "sbd_config", "argument": "SBD_TIMEOUT_ACTION", "check_id": "check1"},
			{"name": "pacemaker_version", "gatherer": "package_version", "argument": "pacemaker", "check_id": "check2"},
			{"name": "corosync_version", "gatherer": "package_version", "argument": "corosync", "check_id": "check3"},
			{"name": "sbd_pcmk_delay_max", "gatherer": "cib", "argument": "//primitive[@type='external/sbd']/instance_attributes/nvpair[@name='pcmk_delay_max']/@value", "check_id": "check4"},
			{"name": "cib_sid", "gatherer": "cib", "argument": "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='SID']/@value", "check_id": "check5"},
			{"name": "corosync_token", "gatherer": "corosync.conf", "argument": "totem.token", "check_id": "check6"},
			{"name": "corosync_join", "gatherer": "corosync.conf", "argument": "totem.join", "check_id": "check6"}
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
					CheckID:  "check1",
				},
				{
					Name:     "sbd_timeout_actions",
					Gatherer: "sbd_config",
					Argument: "SBD_TIMEOUT_ACTION",
					CheckID:  "check1",
				},
			},
			"package_version": {
				{
					Name:     "pacemaker_version",
					Gatherer: "package_version",
					Argument: "pacemaker",
					CheckID:  "check2",
				},
				{
					Name:     "corosync_version",
					Gatherer: "package_version",
					Argument: "corosync",
					CheckID:  "check3",
				},
			},
			"cib": {
				{
					Name:     "sbd_pcmk_delay_max",
					Gatherer: "cib",
					Argument: "//primitive[@type='external/sbd']/instance_attributes/nvpair[@name='pcmk_delay_max']/@value",
					CheckID:  "check4",
				},
				{
					Name:     "cib_sid",
					Gatherer: "cib",
					Argument: "//primitive[@type='SAPHana']/instance_attributes/nvpair[@name='SID']/@value",
					CheckID:  "check5",
				},
			},
			"corosync.conf": {
				{
					Name:     "corosync_token",
					Gatherer: "corosync.conf",
					Argument: "totem.token",
					CheckID:  "check6",
				},
				{
					Name:     "corosync_join",
					Gatherer: "corosync.conf",
					Argument: "totem.join",
					CheckID:  "check6",
				},
			},
		},
	}

	suite.NoError(err)
	suite.Equal(expectedRequests, groupedFactRequsets)
}

func (suite *FactsEngineTestSuite) TestFactsEngineBuildResponse() {
	facts := gatherers.FactsResult{
		ExecutionID: "some-id",
		AgentID:     "some-agent",
		Facts: []gatherers.Fact{
			{
				Name:    "fact1",
				Value:   "1",
				CheckID: "check1",
			},
			{
				Name:    "fact2",
				Value:   "2",
				CheckID: "check2",
			},
		},
	}

	response, err := buildResponse(facts)

	expectedResponse := `{"execution_id":"some-id","agent_id":"some-agent","facts":[{"name":"fact1","value":"1","check_id":"check1"},{"name":"fact2","value":"2","check_id":"check2"}]}`

	suite.NoError(err)
	suite.Equal(expectedResponse, string(response))
}

func (suite *FactsEngineTestSuite) TestFactsEngineGetGatherer() {
	engine := NewFactsEngine("", "")
	g, err := engine.GetGatherer("corosync.conf")

	expectedGatherer := &gatherers.CorosyncConfGatherer{}

	suite.NoError(err)
	suite.Equal(expectedGatherer, g)
}

func (suite *FactsEngineTestSuite) TestFactsEngineGetGathererNotFound() {
	engine := NewFactsEngine("", "")
	_, err := engine.GetGatherer("other")

	suite.EqualError(err, "gatherer other not found")
}

func (suite *FactsEngineTestSuite) TestFactsEngineGetGatherersList() {
	engine := &FactsEngine{ // nolint
		factGatherers: map[string]gatherers.FactGatherer{
			"dummyGatherer1": NewDummyGatherer1(),
			"dummyGatherer2": NewDummyGatherer2(),
			"errorGatherer":  NewErrorGatherer(),
		},
	}

	gatherers := engine.GetGatherersList()

	expectedGatherers := []string{"dummyGatherer1", "dummyGatherer2", "errorGatherer"}

	suite.ElementsMatch(expectedGatherers, gatherers)
}

func (suite *FactsEngineTestSuite) TestFactsEngineGetGatherersListNative() {
	engine := NewFactsEngine("", "")

	gatherers := engine.GetGatherersList()

	expectedGatherers := []string{
		"corosync.conf",
		"corosync-cmapctl",
		"package_version",
		"crm_mon",
		"cibadmin",
		"systemd",
		"sbd_config",
		"verify_password",
	}

	suite.ElementsMatch(expectedGatherers, gatherers)
}

func (suite *FactsEngineTestSuite) TestFactsEnginePrettifyFactResult() {
	fact := gatherers.Fact{
		Name:    "some-fact",
		Value:   1,
		CheckID: "check1",
	}

	prettifiedFact, err := PrettifyFactResult(fact)

	expectedResponse := "{\n  \"name\": \"some-fact\",\n  \"value\": 1,\n  \"check_id\": \"check1\"\n}"

	suite.NoError(err)
	suite.Equal(expectedResponse, prettifiedFact)
}

func (suite *FactsEngineTestSuite) TestFactsEngineMergeGatherers() {
	gatherers1 := map[string]gatherers.FactGatherer{
		"dummy1": NewDummyGatherer1(),
	}
	gatherers2 := map[string]gatherers.FactGatherer{
		"dummy2": NewDummyGatherer2(),
	}

	allGatherers := mergeGatherers(gatherers1, gatherers2)

	expectedGatherers := map[string]gatherers.FactGatherer{
		"dummy1": NewDummyGatherer1(),
		"dummy2": NewDummyGatherer2(),
	}

	suite.Equal(expectedGatherers, allGatherers)
}

func (suite *FactsEngineTestSuite) TestFactsEngineLoadPlugins() {
	pluginsFolder, err := os.MkdirTemp("/tmp/", "test-plugins")
	if err != nil {
		panic(err)
	}
	tmpFile, err := os.CreateTemp(pluginsFolder, "dummy")
	if err != nil {
		panic(err)
	}

	engine := &FactsEngine{ // nolint
		factGatherers: map[string]gatherers.FactGatherer{
			gatherers.CorosyncFactKey: gatherers.NewCorosyncConfGatherer(),
		},
		pluginLoaders: PluginLoaders{
			"rpc": &testPluginLoader{},
		},
	}

	err = engine.LoadPlugins(pluginsFolder)

	pluginName := path.Base(tmpFile.Name())
	expectedGatherers := map[string]gatherers.FactGatherer{
		"corosync.conf": &gatherers.CorosyncConfGatherer{},
		pluginName:      NewDummyGatherer1(),
	}

	suite.NoError(err)
	suite.Equal(expectedGatherers, engine.factGatherers)
}
