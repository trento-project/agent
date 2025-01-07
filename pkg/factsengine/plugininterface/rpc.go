package plugininterface

import (
	"context"
	"net/rpc"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type GathererRPC struct{ client *rpc.Client }

func (g *GathererRPC) RequestGathering(
	ctx context.Context,
	factsRequest []entities.FactRequest,
) ([]entities.Fact, error) {
	var resp []entities.Fact
	var err error

	requestID := uuid.New().String()
	args := GatheringArgs{
		Facts:     factsRequest,
		RequestID: requestID,
	}

	gathering := make(chan error)

	go func() {
		gathering <- g.client.Call("Plugin.ServeGathering", args, &resp)
	}()

	select {
	case <-ctx.Done():
		err = g.client.Call("Plugin.Cancel", requestID, &resp)
		return []entities.Fact{}, err
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
	RequestID string
}

func (s *GathererRPCServer) ServeGathering(args GatheringArgs, resp *[]entities.Fact) error {

	ctx, cancel := context.WithCancel(context.Background())
	if s.cancelMap == nil {
		s.cancelMap = make(map[string]context.CancelFunc)
	}
	s.cancelMap[args.RequestID] = cancel
	defer delete(s.cancelMap, args.RequestID)

	var err error
	*resp, err = s.Impl.Gather(ctx, args.Facts)
	return err
}

func (s *GathererRPCServer) Cancel(requestID string, _ *[]entities.Fact) (_ error) {
	cancel, ok := s.cancelMap[requestID]
	if ok {
		cancel()
		delete(s.cancelMap, requestID)
	} else {
		log.Warnf("Cannot find cancel function for request %s", requestID)
	}

	return nil
}
