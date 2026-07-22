// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

//go:build race

package cmd

import (
	"context"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// gather used to write a shared "cancelled" flag from its signal-handling goroutine and read it
// back right after g.Gather returned, with no synchronization between the two when Gather
// completed on its own rather than via ctx.Done(). It now checks ctx.Err() instead, which is
// safe for concurrent use. This test delivers a real signal after gather() has already returned
// to make sure that goroutine's cancel() call doesn't reintroduce a race; it's only meaningful
// under -race (hence the race build tag), and a no-op signal in a normal run would just be
// unnecessary global side effects.
func TestGatherSignalHandlingDoesNotRaceOnCompletion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("signal delivery semantics differ on windows")
	}
	if _, err := os.Stat("/etc/machine-id"); err != nil {
		t.Skip("requires /etc/machine-id to derive an agent ID, not available in this environment")
	}

	viper.Reset()
	defer viper.Reset()

	viper.Set("gatherer", "dir_scan")
	viper.Set("argument", t.TempDir())
	viper.Set("plugins-folder", t.TempDir())
	viper.Set("log-level", "error")

	gatherCmd := &cobra.Command{}
	gatherCmd.SetContext(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		gather(gatherCmd, nil)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("gather() did not return in time")
	}

	// gather() has returned, but its signal-handling goroutine is still alive, blocked on
	// <-signals. Deliver a real signal now so it performs the write that races with the read
	// gather() already did.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send signal to self: %v", err)
	}

	// Give the signal-handling goroutine time to run before the test (and process) exits, so
	// -race has a chance to observe the write.
	time.Sleep(100 * time.Millisecond)
}
