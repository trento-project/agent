package factsengine

import (
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/factsengine/gatherers"
	"github.com/trento-project/agent/internal/factsengine/plugininterface"
)

func loadPlugins(pluginsFolder string) (map[string]gatherers.FactGatherer, error) {
	pluginFactGatherers := make(map[string]gatherers.FactGatherer)
	log.Debugf("Loading plugins...")

	plugins, err := filepath.Glob(fmt.Sprintf("%s/*", pluginsFolder))
	if err != nil {
		return nil, errors.Wrap(err, "Error running glob operation in the provider plugins folder")
	}

	for _, filePath := range plugins {
		log.Debugf("Loading plugin %s", filePath)
		loadedPlugin, err := loadRPCPlugin(filePath)

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

func loadRPCPlugin(pluginPath string) (gatherers.FactGatherer, error) {
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

func cleanupPlugins() {
	goplugin.CleanupClients()
}
