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

// Only meaningful under -race: delivers a signal after gather() has already returned, to catch
// a race between the signal-handling goroutine's cancel() and the completed call.
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
