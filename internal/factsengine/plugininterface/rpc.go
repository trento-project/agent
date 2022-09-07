package plugininterface

import (
	"net/rpc"

	"github.com/trento-project/agent/internal/factsengine/entities"
)

type GathererRPC struct{ client *rpc.Client }

func (g *GathererRPC) Gather(factsRequest []entities.FactDefinition) ([]entities.FactsGatheredItem, error) {
	var resp []entities.FactsGatheredItem

	err := g.client.Call("Plugin.Gather", factsRequest, &resp)

	return resp, err
}

type GathererRPCServer struct {
	Impl Gatherer
}

func (s *GathererRPCServer) Gather(args []entities.FactDefinition, resp *[]entities.FactsGatheredItem) error {
	var err error
	*resp, err = s.Impl.Gather(args)
	return err
}
