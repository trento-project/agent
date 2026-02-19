package cluster

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/trento-project/agent/pkg/utils"
)

const resourceRefreshedMessage = "got reply (done)"
const idleState = "S_IDLE"

type CmdClient interface {
	GetState(ctx context.Context) (string, error)
	IsHostOnline(ctx context.Context) bool
	IsIdle(ctx context.Context) (bool, error)
	ResourceRefresh(ctx context.Context, resourceID, nodeID string) error
	StartCluster(ctx context.Context) error
	StopCluster(ctx context.Context) error
}

type client struct {
	executor utils.CommandExecutor
	logger   *slog.Logger
}

func NewDefaultCmdClient() CmdClient {
	return NewCmdClient(
		utils.Executor{},
		slog.Default(),
	)
}

func NewCmdClient(executor utils.CommandExecutor, logger *slog.Logger) CmdClient {
	return &client{
		executor: executor,
		logger:   logger,
	}
}

// GetState returns the current state of the cluster using crmadmin command
// Find all existing states here:
// https://github.com/ClusterLabs/pacemaker/blob/main/daemons/controld/controld_fsa.h
func (c *client) GetState(ctx context.Context) (string, error) {
	// Adding a timeout as crmadmin command can hang forever when it is used
	// in a cluster that was recently started and the DC is not selected yet
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	dcNode, err := c.executor.CombinedOutputContext(ctxWithTimeout, "crmadmin", "-qD")
	if err != nil {
		return "", fmt.Errorf("error getting DC node with crmadmin: %w", err)
	}
	state, err := c.executor.CombinedOutputContext(ctxWithTimeout, "crmadmin", "-qS", strings.TrimSpace(string(dcNode)))
	if err != nil {
		return "", fmt.Errorf("error getting cluster state with crmadmin: %w", err)
	}

	return strings.TrimSpace(string(state)), nil
}

func (c *client) IsHostOnline(ctx context.Context) bool {
	output, err := c.executor.CombinedOutputContext(ctx, "crm", "status")
	if err != nil {
		return false
	}

	c.logger.Debug("CRM status output", "output", string(output))

	return true
}

func (c *client) StartCluster(ctx context.Context) error {
	c.logger.Info("Starting CRM cluster")
	output, err := c.executor.CombinedOutputContext(ctx, "crm", "cluster", "start")
	if err != nil {
		return fmt.Errorf("failed to start CRM cluster: %w, output: %s", err, string(output))
	}

	c.logger.Info("CRM cluster started successfully")
	return nil
}

func (c *client) StopCluster(ctx context.Context) error {
	c.logger.Info("Stopping CRM cluster")
	output, err := c.executor.CombinedOutputContext(ctx, "crm", "cluster", "stop")
	if err != nil {
		return fmt.Errorf("failed to stop CRM cluster: %w, output: %s", err, string(output))
	}

	c.logger.Info("CRM cluster stopped successfully")
	return nil
}

func (c *client) IsIdle(ctx context.Context) (bool, error) {
	state, err := c.GetState(ctx)
	if err != nil {
		return false, err
	}

	return state == idleState, nil
}

// ResourceRefresh runs the `crm resource refresh [<rsc>] [<node>]` command.
// https://crmsh.github.io/man-5.0/#cmdhelp.resource.refresh
// The node argument requires the resource beforehand.
// If the given node is not found, the command does not return -1, so the
// std output must be compared to see if it returns a correct value.
func (c *client) ResourceRefresh(ctx context.Context, resourceID, nodeID string) error {
	if nodeID != "" && resourceID == "" {
		return errors.New("nodeID cannot be provided without a resourceID")
	}

	args := []string{"resource", "refresh"}
	if resourceID != "" {
		args = append(args, resourceID)
	}

	if nodeID != "" {
		args = append(args, nodeID)
	}

	c.logger.Info("Refreshing cluster resource", "resourceID", resourceID, "nodeID", nodeID)
	output, err := c.executor.CombinedOutputContext(ctx, "crm", args...)
	if err != nil {
		return fmt.Errorf("failed to refresh resource: %w, output: %s", err, string(output))
	}

	if !strings.Contains(string(output), resourceRefreshedMessage) {
		return fmt.Errorf("failed to refresh resource, unexpected output: %s", string(output))
	}

	c.logger.Info("Cluster resource refreshed successfully")
	return nil
}
