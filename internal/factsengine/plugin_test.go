package factsengine

import (
	"errors"
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

type PluginTestSuite struct {
	suite.Suite
}

func TestPluginTestSuite(t *testing.T) {
	suite.Run(t, new(PluginTestSuite))
}

type testPluginLoader struct{}

func (l *testPluginLoader) Load(pluginPath string) (gatherers.FactGatherer, error) {
	return NewDummyGatherer1(), nil
}

type errorPluginLoader struct{}

func (l *errorPluginLoader) Load(pluginPath string) (gatherers.FactGatherer, error) {
	return nil, errors.New("kaboom")
}

func (suite *PluginTestSuite) TestPluginLoadPlugins() {
	pluginsFolder, err := ioutil.TempDir("/tmp/", "test-plugins")
	if err != nil {
		panic(err)
	}
	plugin1, err := ioutil.TempFile(pluginsFolder, "plugin1")
	if err != nil {
		panic(err)
	}
	plugin2, err := ioutil.TempFile(pluginsFolder, "plugin2")
	if err != nil {
		panic(err)
	}

	loaders := PluginLoaders{
		"rpc": &testPluginLoader{},
	}

	loadedPlugins, err := loadPlugins(loaders, pluginsFolder)

	plugin1Name := path.Base(plugin1.Name())
	plugin2Name := path.Base(plugin2.Name())
	expectedGatherers := map[string]gatherers.FactGatherer{
		plugin1Name: NewDummyGatherer1(),
		plugin2Name: NewDummyGatherer1(),
	}

	suite.NoError(err)
	suite.Equal(expectedGatherers, loadedPlugins)
}

func (suite *PluginTestSuite) TestPluginLoadPluginsError() {
	pluginsFolder, err := ioutil.TempDir("/tmp/", "test-plugins")
	if err != nil {
		panic(err)
	}
	_, err = ioutil.TempFile(pluginsFolder, "plugin")
	if err != nil {
		panic(err)
	}

	loaders := PluginLoaders{
		"rpc": &errorPluginLoader{},
	}

	loadedPlugins, err := loadPlugins(loaders, pluginsFolder)

	expectedGatherers := map[string]gatherers.FactGatherer{}

	suite.NoError(err)
	suite.Equal(expectedGatherers, loadedPlugins)
}
