// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package support_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/support"
)

type RetryTestSuite struct {
	suite.Suite
}

func TestRetry(t *testing.T) {
	suite.Run(t, new(RetryTestSuite))
}

func stringAfterNRetries(n int, value string) func() (string, error) {
	count := 0

	return func() (string, error) {
		count++
		if count < n {
			return "", errors.New("flaky error")
		}

		return value, nil
	}
}

func (suite *RetryTestSuite) TestAsyncExponentialBackoff() {
	const expectedValue = "success"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultCh := support.AsyncExponentialBackoff(
		ctx,
		support.BackoffOptions{
			MaxRetries:   5,
			InitialDelay: 200 * time.Millisecond,
			MaxDelay:     3 * time.Second,
			Factor:       2,
		},
		stringAfterNRetries(3, expectedValue),
	)

	result := <-resultCh

	suite.Require().NoError(result.Err, "Expected no error after retries")
	suite.Equal(expectedValue, result.Result, "Expected successful operation result")
}

func (suite *RetryTestSuite) TestAsyncExponentialBackoffWithCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately to simulate cancellation

	resultCh := support.AsyncExponentialBackoff(
		ctx,
		support.BackoffOptions{
			MaxRetries:   5,
			InitialDelay: 200 * time.Millisecond,
			MaxDelay:     3 * time.Second,
			Factor:       2,
		},
		stringAfterNRetries(3, "should not be reached"),
	)

	result := <-resultCh

	suite.Require().Error(result.Err, "Expected error due to context cancellation")
	suite.Require().EqualError(result.Err, context.Canceled.Error(), "Expected context.Canceled error")
	suite.Empty(result.Result, "Expected empty result due to cancellation")
}

func (suite *RetryTestSuite) TestAsyncExponentialBackoffWithMaxRetries() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultCh := support.AsyncExponentialBackoff(
		ctx,
		support.BackoffOptions{
			MaxRetries:   3,
			InitialDelay: 200 * time.Millisecond,
			MaxDelay:     3 * time.Second,
			Factor:       2,
		},
		stringAfterNRetries(5, "should not be reached"),
	)

	result := <-resultCh

	suite.Require().Error(result.Err, "Expected error after max retries")
	suite.Contains(result.Err.Error(), "operation failed after 3 attempts", "Expected specific error message")
	suite.Empty(result.Result, "Expected empty result due to failure")
}

func (suite *RetryTestSuite) TestAsyncExponentialBackoffWithImmediateSuccess() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultCh := support.AsyncExponentialBackoff(
		ctx,
		support.BackoffOptions{
			MaxRetries:   5,
			InitialDelay: 200 * time.Millisecond,
			MaxDelay:     3 * time.Second,
			Factor:       2,
		},
		func() (bool, error) {
			return true, nil
		},
	)

	result := <-resultCh

	suite.Require().NoError(result.Err, "Expected no error for immediate success")
	suite.True(result.Result, "Expected immediate success result")
}

func (suite *RetryTestSuite) TestAsyncExponentialBackoffOperationFails() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultCh := support.AsyncExponentialBackoff(
		ctx,
		support.BackoffOptions{
			MaxRetries:   1,
			InitialDelay: 200 * time.Millisecond,
			MaxDelay:     3 * time.Second,
			Factor:       2,
		},
		func() (bool, error) {
			return false, errors.New("custom error")
		},
	)

	result := <-resultCh

	suite.Require().Error(result.Err, "Expected error due to operation failure")
	suite.Require().EqualError(result.Err, "operation failed after 1 attempts: custom error")
	suite.Empty(result.Result, "Expected empty result due to operation failure")
}

func TestCalculateDelayExponentialGrowth(t *testing.T) {
	options := support.BackoffOptions{
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
		got := support.CalculateDelay(c.attempt, options)

		if got < 0 {
			t.Errorf("attempt %d: delay must never be negative, got %s", c.attempt, got)
		}

		if got != c.expected {
			t.Errorf("attempt %d: expected delay %s, got %s", c.attempt, c.expected, got)
		}
	}
}

func TestCalculateDelayCapsAtMaxDelay(t *testing.T) {
	options := support.BackoffOptions{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Factor:       2,
	}

	got := support.CalculateDelay(5, options)

	if got != options.MaxDelay {
		t.Errorf("expected delay capped at MaxDelay (%s), got %s", options.MaxDelay, got)
	}
}

func TestCalculateDelayDoesNotOverflowWithLargeAttempt(t *testing.T) {
	options := support.BackoffOptions{
		InitialDelay: time.Second,
		MaxDelay:     time.Minute,
		Factor:       2,
	}

	got := support.CalculateDelay(100, options)

	if got < 0 {
		t.Errorf("delay must never be negative, got %s", got)
	}

	if got != options.MaxDelay {
		t.Errorf("expected delay capped at MaxDelay (%s) for a large attempt count, got %s", options.MaxDelay, got)
	}
}

func TestCalculateDelayFirstAttemptIsZero(t *testing.T) {
	options := support.BackoffOptions{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Factor:       2,
	}

	got := support.CalculateDelay(0, options)

	if got != 0 {
		t.Errorf("expected zero delay for attempt < 1, got %s", got)
	}
}
