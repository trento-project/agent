package operator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/trento-project/agent/internal/core/cluster"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	ClusterMaintenanceChangeOperatorName = "clustermaintenancechange"
	nodeStatePattern                     = "value=(.*)"
	maintenanceOn                        = "on"
	maintenanceOff                       = "off"
)

type clusterMaintenanceScope int

const (
	clusterScope clusterMaintenanceScope = iota
	resourceScope
	nodeScope
)

var nodeStatePatternCompiled = regexp.MustCompile(nodeStatePattern)

type ClusterMaintenanceChangeOption Option[ClusterMaintenanceChange]

type clusterMaintenanceChangeArguments struct {
	maintenance bool
	resourceID  string
	nodeID      string
}

type diffOutput struct {
	Maintenance bool   `json:"maintenance"`
	ResourceID  string `json:"resource_id,omitempty"`
	NodeID      string `json:"node_id,omitempty"`
}

// ClusterMaintenanceChange is an operator responsible for changing cluster maintenance,
// cluster resources or cluster node managed state. `crmsh` is the tool used to apply the write and most of
// the read operations in the cluster.
// The used commands differ if the state to change is the whole cluster, a particular resource or node.
//
// Find some helpful references about maintenance transitions and used commands:
// - https://www.suse.com/c/sles-for-sap-hana-maintenance-procedures-part-1-pre-maintenance-checks/
// nolint:lll
// - https://www.suse.com/c/sles-for-sap-hana-maintenance-procedures-part-2-manual-administrative-tasks-os-reboots-and-updation-of-os-and-hana/
// - https://crmsh.github.io/man-4.6/
// - https://crmsh.github.io/man-4.6/#cmdhelp_root_status
// - https://crmsh.github.io/man-4.6/#cmdhelp_maintenance
// - https://crmsh.github.io/man-4.6/#cmdhelp_resource
// - https://crmsh.github.io/man-4.6/#cmdhelp_node
//
// The operator accepts the next arguments:
// - maintenance (bool): The desired maintenance state for the cluster, resource or node.
//                       If true, the cluster, resource or node are set in maintenance mode.
// - resource_id (string): If given, the operator changes the maintenance state of the resource.
// - node_id (string): If given, the operator changes the maintenance state of the node.
// If resource_id or node_id are not given the operator changes the general maintenance state of the cluster.
// resource_id and node_id mutually exclusive.

//
// # Execution Phases
//
// - PLAN:
//   Check if a pacemaker cluster is present and store the current state.
//
// - COMMIT:
//   Change the cluster, resource or node state if the cluster is in IDLE state.
//   If the maintenance state is removed, the cluster state is refreshed.
//
// - VERIFY:
//   Check if the cluster, resource or node maintenance state has the expected value and
//   store the final state.
//
// - ROLLBACK:
//   Change the cluster, resource or node state to the initial state if the cluster
//   is in IDLE state.

type ClusterMaintenanceChange struct {
	baseOperator
	executor        utils.CommandExecutor
	clusterClient   cluster.CmdClient
	scope           clusterMaintenanceScope
	parsedArguments *clusterMaintenanceChangeArguments
}

func WithCustomClusterMaintenanceExecutor(executor utils.CommandExecutor) ClusterMaintenanceChangeOption {
	return func(o *ClusterMaintenanceChange) {
		o.executor = executor
	}
}

func WithCustomClusterMaintenanceClient(clusterClient cluster.CmdClient) ClusterMaintenanceChangeOption {
	return func(o *ClusterMaintenanceChange) {
		o.clusterClient = clusterClient
	}
}

func NewClusterMaintenanceChange(
	arguments Arguments,
	operationID string,
	options Options[ClusterMaintenanceChange],
) *Executor {
	clusterMaintenance := &ClusterMaintenanceChange{
		baseOperator: newBaseOperator(
			ClusterMaintenanceChangeOperatorName, operationID, arguments, options.BaseOperatorOptions...,
		),
		executor:      utils.Executor{},
		clusterClient: cluster.NewDefaultCmdClient(),
	}

	for _, opt := range options.OperatorOptions {
		opt(clusterMaintenance)
	}

	return &Executor{
		phaser:      clusterMaintenance,
		operationID: operationID,
		logger:      clusterMaintenance.logger,
	}
}

func (c *ClusterMaintenanceChange) plan(ctx context.Context) (bool, error) {
	opArguments, err := parseClusterMaintenanceArguments(c.arguments)
	if err != nil {
		return false, err
	}
	c.parsedArguments = opArguments

	switch {
	case c.parsedArguments.resourceID != "":
		c.scope = resourceScope
	case c.parsedArguments.nodeID != "":
		c.scope = nodeScope
	default:
		c.scope = clusterScope
	}

	// check if a cluster is available and running
	if !c.clusterClient.IsHostOnline(ctx) {
		return false, errors.New("cluster is not runnint on host")
	}

	currentState, err := getMaintenanceState(ctx, c.executor, c.scope, c.parsedArguments)
	if err != nil {
		return false, err
	}

	c.resources[beforeDiffField] = currentState

	if c.resources[beforeDiffField] == c.parsedArguments.maintenance {
		c.logger.Info("maintenance state already set, skipping operation", "state", c.parsedArguments.maintenance)
		c.resources[afterDiffField] = currentState
		return true, nil
	}

	return false, nil
}

func (c *ClusterMaintenanceChange) commit(ctx context.Context) error {
	isIdle, err := c.clusterClient.IsIdle(ctx)
	if err != nil {
		return fmt.Errorf("error checking if cluster is idle: %w", err)
	}

	if !isIdle {
		return errors.New("cluster is not in S_IDLE state")
	}

	// refresh cluster or resource before removing maintenance state
	// in case of node state change, using the command with empty resourceID is OK
	if !c.parsedArguments.maintenance {
		err = c.clusterClient.ResourceRefresh(ctx, c.parsedArguments.resourceID, "")
		if err != nil {
			return fmt.Errorf("error refreshing maintenance state: %w", err)
		}
	}

	err = setMaintenanceState(ctx, c.executor, c.scope, c.parsedArguments.maintenance, c.parsedArguments)
	if err != nil {
		return fmt.Errorf("error updating maintenance state: %w", err)
	}
	return nil
}

func (c *ClusterMaintenanceChange) verify(ctx context.Context) error {
	currentState, err := getMaintenanceState(ctx, c.executor, c.scope, c.parsedArguments)
	if err != nil {
		return err
	}

	if c.parsedArguments.maintenance == currentState {
		c.resources[afterDiffField] = currentState
		return nil
	}

	return fmt.Errorf(
		"verify cluster maintenance failed, the maintenance value %v was not set in commit phase",
		c.parsedArguments.maintenance,
	)
}

func (c *ClusterMaintenanceChange) rollback(ctx context.Context) error {
	isIdle, err := c.clusterClient.IsIdle(ctx)
	if err != nil {
		return fmt.Errorf("error checking if cluster is idle: %w", err)
	}

	if !isIdle {
		return errors.New("cluster is not in S_IDLE state")
	}

	initialState, _ := c.resources[beforeDiffField].(bool)
	err = setMaintenanceState(ctx, c.executor, c.scope, initialState, c.parsedArguments)
	if err != nil {
		return fmt.Errorf("error rolling back maintenance state: %w", err)
	}
	return nil
}

// nolint: dupl
func (c *ClusterMaintenanceChange) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeMaintenance, ok := c.resources[beforeDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid beforeMaintenance value: cannot parse '%s' to bool",
			c.resources[beforeDiffField]))
	}

	afterMaintenance, ok := c.resources[afterDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid afterMaintenance value: cannot parse '%s' to bool",
			c.resources[afterDiffField]))
	}

	beforeDiffOutput := diffOutput{
		Maintenance: beforeMaintenance,
		ResourceID:  c.parsedArguments.resourceID,
		NodeID:      c.parsedArguments.nodeID,
	}
	before, err := json.Marshal(beforeDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling before diff output: %v", err))
	}
	diff["before"] = string(before)

	afterDiffOutput := diffOutput{
		Maintenance: afterMaintenance,
		ResourceID:  c.parsedArguments.resourceID,
		NodeID:      c.parsedArguments.nodeID,
	}
	after, err := json.Marshal(afterDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling after diff output: %v", err))
	}
	diff["after"] = string(after)

	return diff
}

// getMaintanceState returns the current state of the cluster
// Find additional information here:
// https://clusterlabs.org/projects/pacemaker/doc/2.1/Pacemaker_Explained/html/resources.html#resource-meta-attributes
func getMaintenanceState(
	ctx context.Context,
	executor utils.CommandExecutor,
	scope clusterMaintenanceScope,
	args *clusterMaintenanceChangeArguments,
) (bool, error) {
	switch scope {
	case resourceScope:
		{
			// get "maintenance" attribute of the resource. This has preference over is-managed attribute
			output, err := executor.CombinedOutputContext(ctx, "crm", "resource", "meta", args.resourceID, "show", "maintenance")
			if err != nil {
				return false, fmt.Errorf("error getting maintenance attribute: %w", err)
			}

			if !strings.Contains(string(output), "not found") {
				boolValue, err := parseStateOutput(output)
				if err != nil {
					return false, fmt.Errorf("error decoding maintenance attribute: %w", err)
				}

				return boolValue, nil
			}

			// get "is-managed" attribute of the resource
			output, err = executor.CombinedOutputContext(
				ctx, "crm", "resource", "meta", args.resourceID, "show", "is-managed")
			if err != nil {
				return false, fmt.Errorf("error getting is-managed attribute: %w", err)
			}

			// none of maintenance or is-managed attributes found. Defaulting to not in maintenance
			if strings.Contains(string(output), "not found") {
				return false, nil
			}

			boolValue, err := parseStateOutput(output)
			if err != nil {
				return false, fmt.Errorf("error decoding is-managed attribute: %w", err)
			}

			// is-managed has the opposite logic than maintenance attribute
			return !boolValue, nil
		}
	case nodeScope:
		{
			// this command fails if the node is unknown. Check the output to see if the node is recognized
			// possible outputs:
			// maintenance on: scope=nodes  name=maintenance value=true
			// maintenance off: scope=nodes  name=maintenance value=off
			// yes, it returns true/off instead of true/false, on/off...
			// node not found output:
			// Could not map name=node-name to a UUID
			maintenanceMode, err := executor.CombinedOutputContext(
				ctx, "crm", "node", "attribute", args.nodeID, "show", "maintenance")
			if err != nil && strings.Contains(string(maintenanceMode), "Could not map") {
				return false, fmt.Errorf("error getting node maintenance attribute: %w", err)
			}

			values := nodeStatePatternCompiled.FindSubmatch(maintenanceMode)
			if len(values) == 2 && string(values[1]) == "true" {
				return true, nil
			}

			return false, nil
		}
	default:
		{
			maintenanceMode, err := executor.CombinedOutputContext(
				ctx, "crm", "configure", "get_property", "-t", "maintenance-mode")
			if err != nil {
				return false, fmt.Errorf("error getting maintenance-mode: %w", err)
			}

			boolValue, err := parseStateOutput(maintenanceMode)
			if err != nil {
				return false, fmt.Errorf("error decoding maintenance-mode attribute: %w", err)
			}

			return boolValue, nil
		}
	}
}

func setMaintenanceState(
	ctx context.Context,
	executor utils.CommandExecutor,
	scope clusterMaintenanceScope,
	state bool,
	args *clusterMaintenanceChangeArguments,
) error {
	strState := maintenanceOff
	if state {
		strState = maintenanceOn
	}

	switch scope {
	case resourceScope:
		{
			_, err := executor.CombinedOutputContext(ctx, "crm", "maintenance", strState, args.resourceID)
			return err
		}
	case nodeScope:
		{
			if state {
				_, err := executor.CombinedOutputContext(ctx, "crm", "--force", "node", "maintenance", args.nodeID)
				return err
			}
			_, err := executor.CombinedOutputContext(ctx, "crm", "--force", "node", "ready", args.nodeID)
			return err
		}
	default:
		{
			_, err := executor.CombinedOutputContext(ctx, "crm", "maintenance", strState)
			return err
		}
	}
}

// Depending on the queried resource, the crm command might print some "debug" lines
// before returning the actual state of the attribute.
// The actual state is always a boolean value, either 'true' or 'false'
// The debug lines are cleaned up before parsing the final boolean state of the attribute.
// Example output:
// linux # crm resource meta msl_SAPHana_PRD_HDB00 show maintenance
// msl_SAPHana_PRD_HDB00 is active on more than one node, returning the default value for maintenance
// false
func parseStateOutput(output []byte) (bool, error) {
	trimmedString := strings.TrimSpace(string(output))
	if len(trimmedString) == 0 {
		return false, fmt.Errorf("empty command output")
	}

	lines := strings.Split(trimmedString, "\n")
	lastLine := lines[len(lines)-1]

	boolValue, err := strconv.ParseBool(lastLine)
	if err != nil {
		return false, err
	}
	return boolValue, nil
}

func parseClusterMaintenanceArguments(rawArguments Arguments) (*clusterMaintenanceChangeArguments, error) {
	var resourceID, nodeID string

	maintenanceArgument, found := rawArguments["maintenance"]
	if !found {
		return nil, errors.New("argument maintenance not provided, could not use the operator")
	}

	maintenance, ok := maintenanceArgument.(bool)
	if !ok {
		return nil, fmt.Errorf(
			"could not parse maintenance argument as bool, argument provided: %v",
			maintenanceArgument,
		)
	}

	resourceIDArgument, resourceIDfound := rawArguments["resource_id"]
	nodeIDArgument, nodeIDfound := rawArguments["node_id"]

	if resourceIDfound && nodeIDfound {
		return nil, errors.New("resource_id and node_id arguments are mutually exclusive, use only one of them")
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

	return &clusterMaintenanceChangeArguments{
		maintenance: maintenance,
		resourceID:  resourceID,
		nodeID:      nodeID,
	}, nil
}
