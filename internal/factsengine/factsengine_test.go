//nolint:lll
package factsengine

import (
	"fmt"
	"io/ioutil"
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
			Name:  "dummy1",
			Value: "1",
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
			Name:  "dummy2",
			Value: "2",
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
	someID := "someID" //nolint

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

func (suite *FactsEngineTestSuite) TestFactsEngineGatherFactsGathererNotFound() {
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

func (suite *FactsEngineTestSuite) TestFactsEngineGatherFactsErrorGathering() {
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

func (suite *FactsEngineTestSuite) TestFactsEngineParseFactsRequest() {

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

func (suite *FactsEngineTestSuite) TestFactsEngineBuildResponse() {
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

func (suite *FactsEngineTestSuite) TestFactsEnginePrettifyFactResult() {
	fact := gatherers.Fact{
		Name:  "some-fact",
		Value: 1,
	}

	prettifiedFact, err := PrettifyFactResult(fact)

	expectedResponse := "{\n  \"name\": \"some-fact\",\n  \"value\": 1\n}"

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
	pluginsFolder, err := ioutil.TempDir("/tmp/", "test-plugins")
	if err != nil {
		panic(err)
	}
	tmpFile, err := ioutil.TempFile(pluginsFolder, "dummy")
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
