// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package plugininterface

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/trento-project/agent/pkg/factsengine/entities"
)

// slowGatherer blocks until its context is cancelled, mimicking a long-running gathering
// operation that only stops once RequestGathering asks the server to cancel it. It also gives
// up on its own after a short bound, so a test that races Cancel against registration and
// happens to miss the cancellation can't hang forever.
type slowGatherer struct{}

func (slowGatherer) Gather(ctx context.Context, _ []entities.FactRequest) ([]entities.Fact, error) {
	select {
	case <-ctx.Done():
	case <-time.After(500 * time.Millisecond):
	}
	return nil, ctx.Err()
}

// When RequestGathering's caller context is cancelled, it asks the server to cancel the
// in-flight gathering and returns immediately, while a background goroutine is left waiting for
// the original "ServeGathering" RPC call to complete. If the channel it reports on isn't
// buffered, nobody is left to receive once that call finally returns, and the goroutine leaks.
func TestRequestGatheringCancellationDoesNotLeakGoroutine(t *testing.T) {
	clientConn, serverConn := net.Pipe()

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("Plugin", &GathererRPCServer{Impl: slowGatherer{}}))

	serverDone := make(chan struct{})
	go func() {
		server.ServeConn(serverConn)
		close(serverDone)
	}()

	rpcClient := rpc.NewClient(clientConn)
	g := &GathererRPC{client: rpcClient}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// The cancellation path returns whatever the "Plugin.Cancel" RPC call returns (nil on
	// success), not the original ctx.Err() - so RequestGathering itself won't error here.
	_, err := g.RequestGathering(ctx, []entities.FactRequest{{Name: "test", Argument: "test"}})
	require.NoError(t, err)

	// Tear down the RPC plumbing (test scaffolding, not the code under test) so only a genuine
	// leak from RequestGathering's own background goroutine is left for goleak to catch.
	require.NoError(t, rpcClient.Close())
	require.NoError(t, serverConn.Close())
	<-serverDone

	// goleak.Find already retries internally for a short window; polling it from another
	// goroutine (e.g. via require.Eventually) would make that polling goroutine itself show up
	// as "unexpected", so retry it directly in this goroutine instead.
	var leakErr error
	for i := 0; i < 20; i++ {
		leakErr = goleak.Find()
		if leakErr == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.NoError(t, leakErr, "background RPC call goroutine leaked after context cancellation")
}

// ServeGathering and Cancel are separate RPC methods, so net/rpc can - and does, whenever a
// caller cancels while gathering is still in flight - run them concurrently for the same
// server. Both mutate cancelMap; without synchronization that's a concurrent map access, which
// in Go can crash the process with "fatal error: concurrent map writes" even outside of -race.
func TestGathererRPCServerCancelMapConcurrentAccessIsSynchronized(t *testing.T) {
	server := &GathererRPCServer{Impl: slowGatherer{}}

	const requests = 50
	var wg sync.WaitGroup
	wg.Add(requests * 2)

	for i := 0; i < requests; i++ {
		requestID := fmt.Sprintf("req-%d", i)

		go func() {
			defer wg.Done()
			var resp []entities.Fact
			_ = server.ServeGathering(GatheringArgs{RequestID: requestID}, &resp)
		}()

		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond)
			var resp []entities.Fact
			_ = server.Cancel(requestID, &resp)
		}()
	}

	wg.Wait()

	// Every registered request should have been cleaned up by either ServeGathering's own
	// deferred delete or by Cancel - nothing should be left dangling in the map.
	require.Empty(t, server.cancelMap, "cancelMap should be empty once all requests have completed")
}
