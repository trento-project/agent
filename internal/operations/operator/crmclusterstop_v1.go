// CrmClusterStop operator stops a CRM cluster.
//
// # Execution Phases
//
// - PLAN:
//   Checks if the CRM cluster is already offline. If it is, the operation is skipped.
//   If the cluster is online, it checks if the cluster is idle before proceeding.
//   If the cluster is not idle, it returns an error.
//
// - COMMIT:
//   Stops the CRM cluster using the crmClient's StopCluster method.
//
// - VERIFY:
//   Verifies if the CRM cluster is offline after the stop operation, using exponential backoff retries.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the cluster is started again.
//
// # Details
//
// This operator is designed to safely stop a CRM cluster, ensuring that the cluster is only stopped if it is online.
// It uses a retry mechanism with exponential backoff for rollback and verification phases to handle transient failures.
// The operator provides detailed logging for each phase and maintains before/after state for diff reporting.

package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/internal/support"
)

const (
	CrmClusterStopOperatorName = "crmclusterstop"
)

type CrmClusterStop struct {
	baseOperator
	clusterClient cluster.CmdClient
	retryOptions  support.BackoffOptions
}

type CrmClusterStopOption Option[CrmClusterStop]

type CrmClusterStopDiffOutput struct {
	Stopped bool `json:"stopped"`
}

func WithCustomClusterClientStop(clusterClient cluster.CmdClient) CrmClusterStopOption {
	return func(c *CrmClusterStop) {
		c.clusterClient = clusterClient
	}
}

func WithCustomRetryStop(maxRetries int, initialDelay, maxDelay time.Duration, factor int) CrmClusterStopOption {
	return func(c *CrmClusterStop) {
		c.retryOptions = support.BackoffOptions{
			InitialDelay: initialDelay,
			MaxDelay:     maxDelay,
			MaxRetries:   maxRetries,
			Factor:       factor,
		}
	}
}

func NewCrmClusterStop(arguments Arguments,
	operationID string,
	options Options[CrmClusterStop]) *Executor {
	crmClusterStop := &CrmClusterStop{
		baseOperator: newBaseOperator(
			CrmClusterStopOperatorName, operationID, arguments, options.BaseOperatorOptions...,
		),
		clusterClient: cluster.NewDefaultCmdClient(),
		// wait before each execution: 0s, 1.5s, 4.5s, 13.5s, 40.5s
		retryOptions: support.BackoffOptions{
			InitialDelay: 500 * time.Millisecond,
			MaxDelay:     1 * time.Minute,
			MaxRetries:   5,
			Factor:       3,
		},
	}

	for _, opt := range options.OperatorOptions {
		opt(crmClusterStop)
	}

	return &Executor{
		phaser:      crmClusterStop,
		operationID: operationID,
		logger:      crmClusterStop.logger,
	}
}

func (c *CrmClusterStop) plan(ctx context.Context) (bool, error) {
	// check if the cluster is not started.
	isOnline := c.clusterClient.IsHostOnline(ctx)
	c.resources[beforeDiffField] = !isOnline

	if !isOnline {
		c.logger.Info("CRM cluster is not online, skipping stop operation")
		c.resources[afterDiffField] = true
		return true, nil
	}

	return false, nil
}

func (c *CrmClusterStop) commit(ctx context.Context) error {
	// If the cluster is not idle, we cannot stop it safely.
	err := c.ensureIsIdle(ctx)
	if err != nil {
		return fmt.Errorf("cluster is not in IDLE state, cannot stop: %w", err)
	}

	result := <-support.AsyncExponentialBackoff(
		ctx,
		c.retryOptions,
		func() (bool, error) {
			return true, c.clusterClient.StopCluster(ctx)
		},
	)

	if result.Err != nil {
		return fmt.Errorf("error stopping CRM cluster: %w", result.Err)
	}

	return nil
}

func (c *CrmClusterStop) rollback(ctx context.Context) error {
	// We can start the cluster again if it was stopped successfully.
	return c.clusterClient.StartCluster(ctx)
}

func (c *CrmClusterStop) verify(ctx context.Context) error {
	result := <-support.AsyncExponentialBackoff(
		ctx,
		c.retryOptions,
		func() (bool, error) {
			isOnline := c.clusterClient.IsHostOnline(ctx)
			if isOnline {
				return false, fmt.Errorf("CRM cluster is still online, expected offline state")
			}
			return true, nil
		},
	)

	if result.Err != nil {
		return result.Err
	}

	c.resources[afterDiffField] = true
	return nil
}

//	operationDiff needs to be refactored, ignoring duplication issues for now
//
// nolint: dupl
func (c *CrmClusterStop) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeStopped, ok := c.resources[beforeDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid beforeStopped value: cannot parse '%s' to bool",
			c.resources[beforeDiffField]))
	}

	afterStopped, ok := c.resources[afterDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid afterStopped value: cannot parse '%s' to bool",
			c.resources[afterDiffField]))
	}

	beforeDiffOutput := CrmClusterStopDiffOutput{
		Stopped: beforeStopped,
	}
	before, err := json.Marshal(beforeDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling before diff output: %v", err))
	}
	diff["before"] = string(before)

	afterDiffOutput := CrmClusterStopDiffOutput{
		Stopped: afterStopped,
	}
	after, err := json.Marshal(afterDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling after diff output: %v", err))
	}
	diff["after"] = string(after)

	return diff
}

// Ensure the CRM cluster is idle before proceeding with the operation.
// This is a safety check to ensure that the cluster is in a stable state.
// If the cluster is not idle, we will retry until it becomes idle or the maximum retries are reached.
func (c *CrmClusterStop) ensureIsIdle(ctx context.Context) error {
	result := <-support.AsyncExponentialBackoff(
		ctx,
		c.retryOptions,
		func() (bool, error) {
			isIdle, err := c.clusterClient.IsIdle(ctx)
			if err != nil {
				return false, fmt.Errorf("error checking if CRM cluster is idle: %w", err)
			} else if !isIdle {
				return false, fmt.Errorf("CRM cluster is not idle, expected S_IDLE state")
			}
			return true, nil
		},
	)

	return result.Err
}
