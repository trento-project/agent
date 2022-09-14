//nolint:lll
package factsengine

import (
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

func (suite *FactsEngineTestSuite) TestFactsEngineGetGatherer() {
	engine := NewFactsEngine("", "")
	g, err := engine.GetGatherer("corosync.conf")

	expectedGatherer := gatherers.NewDefaultCorosyncConfGatherer()

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
	}

	suite.ElementsMatch(expectedGatherers, gatherers)
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
			gatherers.CorosyncFactKey: gatherers.NewDefaultCorosyncConfGatherer(),
		},
		pluginLoaders: PluginLoaders{
			"rpc": &testPluginLoader{},
		},
	}

	err = engine.LoadPlugins(pluginsFolder)

	pluginName := path.Base(tmpFile.Name())
	expectedGatherers := map[string]gatherers.FactGatherer{
		"corosync.conf": gatherers.NewDefaultCorosyncConfGatherer(),
		pluginName:      NewDummyGatherer1(),
	}

	suite.NoError(err)
	suite.Equal(expectedGatherers, engine.factGatherers)
}
