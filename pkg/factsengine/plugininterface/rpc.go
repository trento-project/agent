// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package plugininterface

import (
	"context"
	"log/slog"
	"net/rpc"
	"sync"

	"github.com/google/uuid"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type GathererRPC struct{ client *rpc.Client }

func (g *GathererRPC) RequestGathering(
	ctx context.Context,
	factsRequest []entities.FactRequest,
) ([]entities.Fact, error) {
	var (
		resp []entities.Fact
		err  error
	)

	requestID := uuid.New().String()
	args := GatheringArgs{
		FactRequests: factsRequest,
		RequestID:    requestID,
	}

	gathering := make(chan error, 1)

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
	mu        sync.Mutex
}

type GatheringArgs struct {
	FactRequests []entities.FactRequest
	RequestID    string
}

func (s *GathererRPCServer) ServeGathering(args GatheringArgs, resp *[]entities.Fact) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.mu.Lock()
	if s.cancelMap == nil {
		s.cancelMap = make(map[string]context.CancelFunc)
	}

	s.cancelMap[args.RequestID] = cancel
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.cancelMap, args.RequestID)
		s.mu.Unlock()
	}()

	var err error

	*resp, err = s.Impl.Gather(ctx, args.FactRequests)

	return err
}

func (s *GathererRPCServer) Cancel(requestID string, _ *[]entities.Fact) (_ error) {
	s.mu.Lock()
	cancel, ok := s.cancelMap[requestID]
	if ok {
		delete(s.cancelMap, requestID)
	}
	s.mu.Unlock()

	if ok {
		cancel()
	} else {
		slog.Warn("Cannot find cancel function for request", "requestID", requestID)
	}

	return nil
}
