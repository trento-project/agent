package gatherers_test

import (
	"errors"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/gatherers/mocks"
)

type PluginTestSuite struct {
	suite.Suite
}

func TestPluginTestSuite(t *testing.T) {
	suite.Run(t, new(PluginTestSuite))
}

type testPluginLoader struct{}

func (l *testPluginLoader) Load(pluginPath string) (gatherers.FactGatherer, error) {
	return &mocks.FactGatherer{}, nil
}

type errorPluginLoader struct{}

func (l *errorPluginLoader) Load(pluginPath string) (gatherers.FactGatherer, error) {
	return nil, errors.New("kaboom")
}

func (suite *PluginTestSuite) TestPluginLoadPlugins() {
	pluginsFolder, err := os.MkdirTemp("/tmp/", "test-plugins")
	if err != nil {
		panic(err)
	}
	plugin1, err := os.CreateTemp(pluginsFolder, "plugin1")
	if err != nil {
		panic(err)
	}
	plugin2, err := os.CreateTemp(pluginsFolder, "plugin2")
	if err != nil {
		panic(err)
	}

	loaders := gatherers.PluginLoaders{
		"rpc": &testPluginLoader{},
	}

	loadedPlugins, err := gatherers.GetGatherersFromPlugins(loaders, pluginsFolder)

	plugin1Name := path.Base(plugin1.Name())
	plugin2Name := path.Base(plugin2.Name())

	expectedGatherers := gatherers.Tree{
		plugin1Name: map[string]gatherers.FactGatherer{
			"v1": &mocks.FactGatherer{},
		},
		plugin2Name: map[string]gatherers.FactGatherer{
			"v1": &mocks.FactGatherer{},
		},
	}

	suite.NoError(err)
	suite.Equal(expectedGatherers, loadedPlugins)
}

func (suite *PluginTestSuite) TestPluginLoadPluginsError() {
	pluginsFolder, err := os.MkdirTemp("/tmp/", "test-plugins")
	if err != nil {
		panic(err)
	}
	_, err = os.CreateTemp(pluginsFolder, "plugin")
	if err != nil {
		panic(err)
	}

	loaders := gatherers.PluginLoaders{
		"rpc": &errorPluginLoader{},
	}

	loadedPlugins, err := gatherers.GetGatherersFromPlugins(loaders, pluginsFolder)

	expectedGatherers := gatherers.Tree{}

	suite.NoError(err)
	suite.Equal(expectedGatherers, loadedPlugins)
}
