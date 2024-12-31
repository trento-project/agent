package gatherers

import (
	"context"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/factsengine/plugininterface"
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

	pluginClient, ok := raw.(plugininterface.GathererRPC)
	if !ok {
		return nil, errors.Wrap(err, "Error asserting Gatherer type")
	}

	p := &PluggedGatherer{
		pluginClient: pluginClient,
	}

	return p, nil
}

type PluggedGatherer struct {
	pluginClient plugininterface.GathererRPC
}

func (g *PluggedGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	return g.pluginClient.RequestGathering(ctx, factsRequests)
}
