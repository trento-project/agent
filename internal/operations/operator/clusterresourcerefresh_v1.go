// ClusterResourceRefresh operator refreshes the cluster resources.
//
// Find some helpful references about the used command here:
// - https://crmsh.github.io/man-5.0/#cmdhelp.resource.refresh
//
// The operator accepts the following optional arguments:
// - resource_id (string): The ID of a specific resource to refresh.
// - node_id (string): The ID of a specific node where the resource should be refreshed.
//                     This can only be provided if `resource_id` is also specified.
//
// If no arguments are provided, all resources in the cluster are refreshed.
//
// # Execution Phases
//
// - PLAN:
//   Checks if the cluster is available and in an IDLE state. If not, the operation fails.
//
// - COMMIT:
//   Refreshes the cluster resources using `crm resource refresh`.
//
// - VERIFY:
//   This phase is a no-op as there is no persistent state to verify for a refresh.
//
// - ROLLBACK:
//   This phase is a no-op as a refresh operation cannot be rolled back.
//
// # Details
//
// This operator is designed to safely refresh cluster resources, ensuring that the
// cluster is in a stable (IDLE) state before proceeding.

package operator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/trento-project/agent/internal/core/cluster"
)

const (
	ClusterResourceRefreshOperatorName = "clusterresourcerefresh"
)

type clusterResourceRefreshArguments struct {
	resourceID string
	nodeID     string
}

type ClusterResourceRefresh struct {
	baseOperator
	clusterClient   cluster.CmdClient
	parsedArguments *clusterResourceRefreshArguments
}

type ClusterResourceRefreshOption Option[ClusterResourceRefresh]

type clusterRefreshDiffOutput struct {
	Refreshed  bool   `json:"refreshed"`
	ResourceID string `json:"resource_id,omitempty"`
	NodeID     string `json:"node_id,omitempty"`
}

func WithCustomClusterResourceRefreshClient(clusterClient cluster.CmdClient) ClusterResourceRefreshOption {
	return func(o *ClusterResourceRefresh) {
		o.clusterClient = clusterClient
	}
}

func NewClusterResourceRefresh(
	arguments Arguments,
	operationID string,
	options Options[ClusterResourceRefresh],
) *Executor {
	clusterRefresh := &ClusterResourceRefresh{
		baseOperator: newBaseOperator(
			ClusterResourceRefreshOperatorName, operationID, arguments, options.BaseOperatorOptions...,
		),
		clusterClient: cluster.NewDefaultCmdClient(),
	}

	for _, opt := range options.OperatorOptions {
		opt(clusterRefresh)
	}

	return &Executor{
		phaser:      clusterRefresh,
		operationID: operationID,
		logger:      clusterRefresh.logger,
	}
}

func (c *ClusterResourceRefresh) plan(ctx context.Context) (bool, error) {
	opArguments, err := parseClusterResourceRefreshArguments(c.arguments)
	if err != nil {
		return false, err
	}
	c.parsedArguments = opArguments

	// check if a cluster is available and running
	if !c.clusterClient.IsHostOnline(ctx) {
		return false, errors.New("cluster is not runnint on host")
	}

	c.resources[beforeDiffField] = false
	return false, nil
}

func (c *ClusterResourceRefresh) commit(ctx context.Context) error {
	isIdle, err := c.clusterClient.IsIdle(ctx)
	if err != nil {
		return fmt.Errorf("error checking if cluster is idle: %w", err)
	}
	if !isIdle {
		return fmt.Errorf("cluster is not in S_IDLE state")
	}

	return c.clusterClient.ResourceRefresh(ctx, c.parsedArguments.resourceID, c.parsedArguments.nodeID)
}

func (c *ClusterResourceRefresh) verify(_ context.Context) error {
	// A refresh operation is an action that doesn't change a verifiable state.
	// If commit was successful, we consider it done.
	c.logger.Debug("Verify is not applicable for cluster refresh operation.")
	c.resources[afterDiffField] = true
	return nil
}

func (c *ClusterResourceRefresh) rollback(_ context.Context) error {
	// There is no rollback for a refresh operation.
	c.logger.Info("Rollback is not applicable for cluster refresh operation.")
	return nil
}

// nolint: dupl
func (c *ClusterResourceRefresh) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeRefreshed, ok := c.resources[beforeDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid beforeRefreshed value: cannot parse '%s' to bool",
			c.resources[beforeDiffField]))
	}

	afterRefreshed, ok := c.resources[afterDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid afterRefreshed value: cannot parse '%s' to bool",
			c.resources[afterDiffField]))
	}

	beforeDiffOutput := clusterRefreshDiffOutput{
		Refreshed:  beforeRefreshed,
		ResourceID: c.parsedArguments.resourceID,
		NodeID:     c.parsedArguments.nodeID,
	}
	before, err := json.Marshal(beforeDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling before diff output: %v", err))
	}
	diff["before"] = string(before)

	afterDiffOutput := clusterRefreshDiffOutput{
		Refreshed:  afterRefreshed,
		ResourceID: c.parsedArguments.resourceID,
		NodeID:     c.parsedArguments.nodeID,
	}
	after, err := json.Marshal(afterDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling after diff output: %v", err))
	}
	diff["after"] = string(after)

	return diff
}

func parseClusterResourceRefreshArguments(rawArguments Arguments) (*clusterResourceRefreshArguments, error) {
	var resourceID, nodeID string
	var ok bool

	resourceIDArgument, resourceIDfound := rawArguments["resource_id"]
	nodeIDArgument, nodeIDfound := rawArguments["node_id"]

	if !resourceIDfound && nodeIDfound {
		return nil, errors.New("node_id cannot be provided without a resource_id")
	}

	if resourceIDfound {
		resourceID, ok = resourceIDArgument.(string)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse resource_id argument as string, argument provided: %v",
				resourceIDArgument,
			)
		}
	}

	if nodeIDfound {
		nodeID, ok = nodeIDArgument.(string)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse node_id argument as string, argument provided: %v",
				nodeIDArgument,
			)
		}
	}

	return &clusterResourceRefreshArguments{resourceID: resourceID, nodeID: nodeID}, nil
}
