package plugininterface

import (
	"context"
	"net/rpc"

	"github.com/google/uuid"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type GathererRPC struct{ client *rpc.Client }

func (g *GathererRPC) RequestGathering(ctx context.Context, factsRequest []entities.FactRequest) ([]entities.Fact, error) {
	var resp []entities.Fact
	var err error

	requestId := uuid.New().String()
	args := GatheringArgs{
		Facts:     factsRequest,
		RequestId: requestId,
	}

	gathering := make(chan error)

	go func() {
		gathering <- g.client.Call("Plugin.ServeGathering", args, &resp)
		close(gathering)
	}()

	select {
	case <-ctx.Done():
		err = g.client.Call("Plugin.Cancel", requestId, nil)
		return nil, err
	case err = <-gathering:
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

type GathererRPCServer struct {
	Impl      Gatherer
	cancelMap map[string]context.CancelFunc
}

type GatheringArgs struct {
	Facts     []entities.FactRequest
	RequestId string
}

func (s *GathererRPCServer) ServeGathering(args GatheringArgs, resp *[]entities.Fact) error {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	if s.cancelMap == nil {
		s.cancelMap = make(map[string]context.CancelFunc)
	}
	s.cancelMap[args.RequestId] = cancel
	defer delete(s.cancelMap, args.RequestId)

	*resp, err = s.Impl.Gather(ctx, args.Facts)
	return err
}

func (s *GathererRPCServer) Cancel(requestId string) error {
	cancel, ok := s.cancelMap[requestId]
	if ok {
		cancel()
		delete(s.cancelMap, requestId)
	}
	return nil
}
