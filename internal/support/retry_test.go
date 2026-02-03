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

	suite.NoError(result.Err, "Expected no error after retries")
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

	suite.Error(result.Err, "Expected error due to context cancellation")
	suite.EqualError(result.Err, context.Canceled.Error(), "Expected context.Canceled error")
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

	suite.Error(result.Err, "Expected error after max retries")
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

	suite.NoError(result.Err, "Expected no error for immediate success")
	suite.Equal(true, result.Result, "Expected immediate success result")
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

	suite.Error(result.Err, "Expected error due to operation failure")
	suite.EqualError(result.Err, "operation failed after 1 attempts: custom error")
	suite.Empty(result.Result, "Expected empty result due to operation failure")
}
