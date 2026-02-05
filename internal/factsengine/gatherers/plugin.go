package gatherers

import (
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"strings"

	goplugin "github.com/hashicorp/go-plugin"
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
	slog.Debug("Loading plugins...")

	plugins, err := filepath.Glob(fmt.Sprintf("%s/*", pluginsFolder))
	if err != nil {
		return nil, fmt.Errorf("Error running glob operation in the provider plugins folder: %w", err)
	}

	for _, filePath := range plugins {
		slog.Debug("Loading plugin", "filePath", filePath)
		// Only RPC is available by now
		// Using a map already to have an easy way to expand if needed
		// A detecType function should be added in this case
		loadedPlugin, err := loaders["rpc"].Load(filePath)

		if err != nil {
			slog.Warn("Error loading plugin", "filePath", filePath, "error", err)
			continue
		}

		name := path.Base(filePath)
		name = strings.TrimSuffix(name, path.Ext(name))

		pluginFactGatherers[name] = map[string]FactGatherer{
			defaultPluginVersion: loadedPlugin,
		}
		slog.Debug("Plugin loaded properly", "filePath", filePath)
	}

	return pluginFactGatherers, nil
}

func CleanupPlugins() {
	goplugin.CleanupClients()
}
