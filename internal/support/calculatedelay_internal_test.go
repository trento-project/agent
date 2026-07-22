// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package support

import (
	"testing"
	"time"
)

// calculateDelay is expected to grow the delay exponentially: InitialDelay * Factor^(attempt-1).
// A previous implementation used `Factor^attempt-1` where `^` is Go's XOR operator rather than
// exponentiation, producing wrong (and sometimes negative) delays instead of real backoff.
func TestCalculateDelayExponentialGrowth(t *testing.T) {
	options := BackoffOptions{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Factor:       2,
	}

	cases := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 400 * time.Millisecond},
		{4, 800 * time.Millisecond},
		{5, 1600 * time.Millisecond},
	}

	for _, c := range cases {
		got := calculateDelay(c.attempt, options)

		if got < 0 {
			t.Errorf("attempt %d: delay must never be negative, got %s", c.attempt, got)
		}

		if got != c.expected {
			t.Errorf("attempt %d: expected delay %s, got %s", c.attempt, c.expected, got)
		}
	}
}

func TestCalculateDelayCapsAtMaxDelay(t *testing.T) {
	options := BackoffOptions{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Factor:       2,
	}

	got := calculateDelay(5, options)

	if got != options.MaxDelay {
		t.Errorf("expected delay capped at MaxDelay (%s), got %s", options.MaxDelay, got)
	}
}

// A large attempt count (or Factor) can overflow the exponential multiplier; the delay must
// still come out capped at MaxDelay instead of wrapping around to something bogus or negative.
func TestCalculateDelayDoesNotOverflowWithLargeAttempt(t *testing.T) {
	options := BackoffOptions{
		InitialDelay: time.Second,
		MaxDelay:     time.Minute,
		Factor:       2,
	}

	got := calculateDelay(100, options)

	if got < 0 {
		t.Errorf("delay must never be negative, got %s", got)
	}

	if got != options.MaxDelay {
		t.Errorf("expected delay capped at MaxDelay (%s) for a large attempt count, got %s", options.MaxDelay, got)
	}
}

func TestCalculateDelayFirstAttemptIsZero(t *testing.T) {
	options := BackoffOptions{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Factor:       2,
	}

	got := calculateDelay(0, options)

	if got != 0 {
		t.Errorf("expected zero delay for attempt < 1, got %s", got)
	}
}
