package gatherers

import (
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/trento-project/agent/internal/factsengine/plugininterface"
)

type RPCPluginLoader struct{}

func (l *RPCPluginLoader) Load(pluginPath string) (FactGatherer, error) {
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
