package gatherers

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const defaultPluginVersion = "v1"

type PluginLoader interface {
	Load(pluginPath string) (FactGatherer, error)
}

type PluginLoaders map[string]PluginLoader

func GetGatherersFromPlugins(
	loaders PluginLoaders,
	pluginsFolder string,
) (FactGatherersTree, error) {
	pluginFactGatherers := make(FactGatherersTree)
	log.Debugf("Loading plugins...")

	plugins, err := filepath.Glob(fmt.Sprintf("%s/*", pluginsFolder))
	if err != nil {
		return nil, errors.Wrap(err, "Error running glob operation in the provider plugins folder")
	}

	for _, filePath := range plugins {
		log.Debugf("Loading plugin %s", filePath)
		// Only RPC is available by now
		// Using a map already to have an easy way to expand if needed
		// A detecType function should be added in this case
		loadedPlugin, err := loaders["rpc"].Load(filePath)

		if err != nil {
			log.Warnf("Error loading plugin %s: %s", filePath, err)
			continue
		}

		name := path.Base(filePath)
		name = strings.TrimSuffix(name, path.Ext(name))

		pluginFactGatherers[name] = map[string]FactGatherer{
			defaultPluginVersion: loadedPlugin,
		}
		log.Debugf("Plugin %s loaded properly", filePath)
	}

	return pluginFactGatherers, nil
}

func CleanupPlugins() {
	goplugin.CleanupClients()
}
