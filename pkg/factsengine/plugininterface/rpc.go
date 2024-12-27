package plugininterface

import (
	"context"
	"net/rpc"

	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type GathererRPC struct{ client *rpc.Client }

func (g *GathererRPC) Gather(_ context.Context, factsRequest []entities.FactRequest) ([]entities.Fact, error) {
	var resp []entities.Fact

	err := g.client.Call("Plugin.Gather", factsRequest, &resp)

	return resp, err
}

type GathererRPCServer struct {
	Impl Gatherer
}

func (s *GathererRPCServer) Gather(ctx context.Context, args []entities.FactRequest, resp *[]entities.Fact) error {
	var err error
	*resp, err = s.Impl.Gather(ctx, args)
	return err
}
