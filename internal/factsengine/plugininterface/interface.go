package plugininterface

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"

	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

// Gatherer is the interface exposed as a plugin.
type Gatherer interface {
	Gather([]gatherers.FactRequest) ([]gatherers.Fact, error)
}

// This is the implementation of plugin.Plugin
type GathererPlugin struct {
	// Impl Injection
	Impl Gatherer
}

func (p *GathererPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &GathererRPCServer{Impl: p.Impl}, nil
}

func (GathererPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &GathererRPC{client: c}, nil
}
