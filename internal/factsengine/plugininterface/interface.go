package plugininterface

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/trento-project/agent/internal/factsengine/entities"
)

// Gatherer is the interface exposed as a plugin.
type Gatherer interface {
	Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error)
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
