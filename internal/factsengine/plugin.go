package factsengine

import (
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/plugininterface"
)

type PluginLoaders map[string]PluginLoader

type PluginLoader interface {
	Load(pluginPath string) (gatherers.FactGatherer, error)
}

func NewPluginLoaders() PluginLoaders {
	// Using a map to make it potentially extensible in the future with other plugin types
	return PluginLoaders{
		"rpc": &RPCPluginLoader{},
	}
}

type RPCPluginLoader struct{}

func (l *RPCPluginLoader) Load(pluginPath string) (gatherers.FactGatherer, error) {
	pluginMap := map[string]goplugin.Plugin{
		"gatherer": &plugininterface.GathererPlugin{Impl: nil},
	}

	handshakeConfig := goplugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "TRENTO_PLUGIN",
		MagicCookieValue: "gatherer",
	}

	client := goplugin.NewClient(&goplugin.ClientConfig{ // nolint
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(pluginPath),
		Managed:         true,
		AllowedProtocols: []goplugin.Protocol{
			goplugin.ProtocolNetRPC,
		},
		Logger: hclog.Default(),
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, errors.Wrap(err, "Error starting the rpc client")
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("gatherer")
	if err != nil {
		return nil, errors.Wrap(err, "Error dispensing plugin")
	}

	g, ok := raw.(plugininterface.Gatherer)
	if !ok {
		return nil, errors.Wrap(err, "Error asserting Gatherer type")
	}

	return g, nil
}

func GetGatherersFromPlugins(loaders PluginLoaders, pluginsFolder string) (map[string]gatherers.FactGatherer, error) {
	pluginFactGatherers := make(map[string]gatherers.FactGatherer)
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
		pluginFactGatherers[name] = loadedPlugin
		log.Debugf("Plugin %s loaded properly", filePath)
	}

	return pluginFactGatherers, nil
}

func CleanupPlugins() {
	goplugin.CleanupClients()
}
