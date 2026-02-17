// CrmClusterStart operator starts a CRM cluster.
//
// # Execution Phases
//
// - PLAN:
//   Checks if the CRM cluster is already online. If it is, the operation is skipped.
//   If the cluster is offline, it checks if the cluster is idle before proceeding.
//   If the cluster is not idle, it returns an error.
//
// - COMMIT:
//   Starts the CRM cluster using the crmClient's StartCluster method.
//
// - VERIFY:
//   Verifies if the CRM cluster is online after the start operation, using exponential backoff retries.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the cluster is stopped using exponential backoff retries.
//
// # Details
//
// This operator is designed to safely start a CRM cluster, ensuring that the cluster is only started if it is offline.
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
	CrmClusterStartOperatorName = "crmclusterstart"
)

type CrmClusterStart struct {
	baseOperator
	clusterClient cluster.CmdClient
	retryOptions  support.BackoffOptions
}

type CrmClusterStartOption Option[CrmClusterStart]

type crmClusterStartDiffOutput struct {
	Started bool `json:"started"`
}

func WithCustomClusterClient(clusterClient cluster.CmdClient) CrmClusterStartOption {
	return func(c *CrmClusterStart) {
		c.clusterClient = clusterClient
	}
}

func WithCustomRetry(maxRetries int, initialDelay, maxDelay time.Duration, factor int) CrmClusterStartOption {
	return func(c *CrmClusterStart) {
		c.retryOptions = support.BackoffOptions{
			InitialDelay: initialDelay,
			MaxDelay:     maxDelay,
			MaxRetries:   maxRetries,
			Factor:       factor,
		}
	}
}

func NewCrmClusterStart(arguments Arguments,
	operationID string,
	options Options[CrmClusterStart]) *Executor {
	crmClusterStart := &CrmClusterStart{
		baseOperator: newBaseOperator(
			CrmClusterStartOperatorName, operationID, arguments, options.BaseOperatorOptions...,
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
		opt(crmClusterStart)
	}

	return &Executor{
		phaser:      crmClusterStart,
		operationID: operationID,
		logger:      crmClusterStart.logger,
	}
}

func (c *CrmClusterStart) plan(ctx context.Context) (bool, error) {
	// check if the cluster is already started.
	isOnline := c.clusterClient.IsHostOnline(ctx)
	c.resources[beforeDiffField] = isOnline

	if isOnline {
		c.logger.Info("CRM cluster is already online, skipping start operation")
		c.resources[afterDiffField] = true
		return true, nil
	}

	return false, nil
}

func (c *CrmClusterStart) commit(ctx context.Context) error {
	err := c.clusterClient.StartCluster(ctx)
	if err != nil {
		return fmt.Errorf("error starting CRM cluster: %w", err)
	}

	return nil
}

func (c *CrmClusterStart) rollback(ctx context.Context) error {
	// If the cluster is not idle, we cannot rollback the start operation.
	err := c.ensureIsIdle(ctx)
	if err != nil {
		return fmt.Errorf("cluster is not in IDLE state, cannot rollback: %w", err)
	}

	result := <-support.AsyncExponentialBackoff(
		ctx,
		c.retryOptions,
		func() (bool, error) {
			return true, c.clusterClient.StopCluster(ctx)
		},
	)

	if result.Err != nil {
		return fmt.Errorf("error rolling back CRM cluster start: %w", result.Err)
	}

	return nil
}

func (c *CrmClusterStart) verify(ctx context.Context) error {
	result := <-support.AsyncExponentialBackoff(
		ctx,
		c.retryOptions,
		func() (bool, error) {
			isOnline := c.clusterClient.IsHostOnline(ctx)
			if !isOnline {
				return false, fmt.Errorf("CRM cluster is not online, expected online state")
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
func (c *CrmClusterStart) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeStarted, ok := c.resources[beforeDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid beforeStarted value: cannot parse '%s' to bool",
			c.resources[beforeDiffField]))
	}
	afterStarted, ok := c.resources[afterDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid afterStarted value: cannot parse '%s' to bool",
			c.resources[afterDiffField]))
	}

	beforeDiffOutput := crmClusterStartDiffOutput{
		Started: beforeStarted,
	}
	before, err := json.Marshal(beforeDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling before diff output: %v", err))
	}
	diff["before"] = string(before)

	afterDiffOutput := crmClusterStartDiffOutput{
		Started: afterStarted,
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
func (c *CrmClusterStart) ensureIsIdle(ctx context.Context) error {
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
