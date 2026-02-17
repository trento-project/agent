package support

import (
	"context"
	"fmt"
	"time"
)

type BackoffOptions struct {
	// MaxRetries is the maximum number of retries before giving up. Set 1 for no retries.
	MaxRetries int
	// InitialDelay is the initial delay before the first retry. Set 0 for no delay.
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration
	// Factor is the multiplier for the backoff delay. Set 1 for linear backoff.
	Factor int
}

// AsyncExponentialBackoff retries an operation asynchronously with exponential backoff.
// It returns a channel that will receive the result or error after the operation succeeds or all retries
// are exhausted.
// The operation is expected to return a value of type T and an error.
// The function will retry up to maxRetries times, starting with initialDelay and doubling the delay
// each time, up to maxDelay.
// If the context is canceled, it will stop retrying and return the context's error.
// If the operation fails after all retries, it will return an error indicating the failure.
func AsyncExponentialBackoff[T any](
	ctx context.Context,
	options BackoffOptions,
	operation func() (T, error),
) <-chan struct {
	Result T
	Err    error
} {
	result := make(chan struct {
		Result T
		Err    error
	}, 1) // buffered so sender does not block

	go func() {
		var zero T
		defer close(result)

		for attempt := 1; attempt <= options.MaxRetries; attempt++ {
			select {
			case <-ctx.Done():
				result <- struct {
					Result T
					Err    error
				}{zero, ctx.Err()}
				return
			default:
			}
			res, err := operation()
			if err == nil {
				result <- struct {
					Result T
					Err    error
				}{res, nil}
				return
			}
			if attempt == options.MaxRetries {
				result <- struct {
					Result T
					Err    error
				}{zero, fmt.Errorf("operation failed after %d attempts: %w", options.MaxRetries, err)}
				return
			}

			delay := calculateDelay(attempt, options)

			select {
			case <-ctx.Done():
				result <- struct {
					Result T
					Err    error
				}{zero, ctx.Err()}
				return
			case <-time.After(delay):
			}

		}
	}()
	return result
}

func calculateDelay(attempt int, options BackoffOptions) time.Duration {
	if attempt < 1 {
		return 0
	}
	delay := options.InitialDelay * time.Duration(options.Factor^attempt-1)
	if delay > options.MaxDelay {
		return options.MaxDelay
	}
	return delay
}
