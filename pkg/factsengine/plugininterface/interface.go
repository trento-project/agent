package plugininterface

import (
	"encoding/gob"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

// TODO: move this to a common place in the pkg folder
// This is needed by the plugin system to be able to serialize the FactValue type
func init() {
	gob.Register(&entities.FactValueInt{})
	gob.Register(&entities.FactValueFloat{})
	gob.Register(&entities.FactValueString{})
	gob.Register(&entities.FactValueBool{})
	gob.Register(&entities.FactValueList{})
	gob.Register(&entities.FactValueMap{})
}

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
