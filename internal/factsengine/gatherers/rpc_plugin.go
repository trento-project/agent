package gatherers

import (
	"context"
	"fmt"
	"os"
	"os/exec"

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
		SyncStdout: os.Stdout,
		SyncStderr: os.Stderr,
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("Error starting the rpc client: %w", err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("gatherer")
	if err != nil {
		return nil, fmt.Errorf("Error dispensing plugin: %w", err)
	}

	pluginClient, ok := raw.(plugininterface.GathererRPC)
	if !ok {
		return nil, fmt.Errorf("Error asserting Gatherer type: %w", err)
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
