package plugininterface

import (
	"net/rpc"

	"github.com/trento-project/agent/internal/factsengine/gatherers"
)

type GathererRPC struct{ client *rpc.Client }

func (g *GathererRPC) Gather(factsRequest []gatherers.FactRequest) ([]gatherers.Fact, error) {
	var resp []gatherers.Fact

	err := g.client.Call("Plugin.Gather", factsRequest, &resp)

	return resp, err
}

type GathererRPCServer struct {
	Impl Gatherer
}

func (s *GathererRPCServer) Gather(args []gatherers.FactRequest, resp *[]gatherers.Fact) error {
	var err error
	*resp, err = s.Impl.Gather(args)
	return err
}
