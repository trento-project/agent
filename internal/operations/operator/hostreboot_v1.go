// HostReboot operator schedules a host reboot after a specified delay.
//
// # Execution Phases
//
// - PLAN:
//   Checks if there is already a scheduled reboot using systemd D-Bus.
//   If a reboot is already scheduled, the operation is skipped.
//
// - COMMIT:
//   Executes the `shutdown -r +1` command to schedule the reboot.
//
// - VERIFY:
//   Verifies that the reboot has been correctly scheduled by checking systemd D-Bus again.
//
// - ROLLBACK:
//   Cancels the scheduled reboot using `shutdown -c` command.
//
// # Details
//
// This operator is designed to safely schedule host reboots, ensuring that multiple
// reboot schedules are not created. It uses systemd D-Bus to check for existing
// scheduled shutdowns and verify the reboot scheduling.

package operator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/trento-project/agent/internal/core/dbus"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	HostRebootOperatorName = "hostreboot"
)

type HostReboot struct {
	baseOperator
	executor        utils.CommandExecutor
	dbusConstructor func(ctx context.Context) (dbus.Connector, error)
}

type HostRebootOption Option[HostReboot]

type hostRebootDiffOutput struct {
	Scheduled bool `json:"scheduled"`
}

func WithCustomHostRebootExecutor(executor utils.CommandExecutor) HostRebootOption {
	return func(o *HostReboot) {
		o.executor = executor
	}
}

func WithCustomDbusConstructor(constructor func(ctx context.Context) (dbus.Connector, error)) HostRebootOption {
	return func(o *HostReboot) {
		o.dbusConstructor = constructor
	}
}

func WithStaticDbusConnector(connector dbus.Connector) HostRebootOption {
	return func(o *HostReboot) {
		o.dbusConstructor = func(_ context.Context) (dbus.Connector, error) {
			return connector, nil
		}
	}
}

func defaultDbusConstructor(ctx context.Context) (dbus.Connector, error) {
	connector, err := dbus.NewConnector(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create D-Bus connector: %w", err)
	}
	return connector, nil
}

func NewHostReboot(arguments Arguments,
	operationID string,
	options Options[HostReboot]) *Executor {
	hostReboot := &HostReboot{
		baseOperator: newBaseOperator(
			HostRebootOperatorName, operationID, arguments, options.BaseOperatorOptions...,
		),
		executor:        utils.Executor{},
		dbusConstructor: defaultDbusConstructor,
	}

	for _, opt := range options.OperatorOptions {
		opt(hostReboot)
	}

	return &Executor{
		phaser:      hostReboot,
		operationID: operationID,
		logger:      hostReboot.logger,
	}
}

func (h *HostReboot) plan(ctx context.Context) (bool, error) {
	// Check if there is already a scheduled reboot
	isScheduled, err := h.isRebootScheduled(ctx)
	if err != nil {
		return false, fmt.Errorf("error checking if reboot is scheduled: %w", err)
	}

	h.resources[beforeDiffField] = isScheduled

	if isScheduled {
		h.resources[afterDiffField] = true
		return true, nil
	}

	return false, nil
}

func (h *HostReboot) commit(ctx context.Context) error {
	_, err := h.executor.CombinedOutputContext(ctx, "shutdown", "-r", "+1", "Host reboot scheduled by automation")

	return err
}

func (h *HostReboot) rollback(ctx context.Context) error {
	_, err := h.executor.CombinedOutputContext(ctx, "shutdown", "-c")
	return err
}

func (h *HostReboot) verify(ctx context.Context) error {
	// Verify that the reboot has been scheduled
	isScheduled, err := h.isRebootScheduled(ctx)
	if err != nil {
		return fmt.Errorf("error verifying reboot scheduling: %w", err)
	}

	if !isScheduled {
		return fmt.Errorf("reboot verification failed: no scheduled reboot found")
	}

	h.resources[afterDiffField] = true
	return nil
}

func (h *HostReboot) operationDiff(_ context.Context) map[string]any {
	diff := make(map[string]any)

	beforeScheduled, ok := h.resources[beforeDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid beforeScheduled value: cannot parse '%s' to bool",
			h.resources[beforeDiffField]))
	}

	beforeDiffOutput := hostRebootDiffOutput{
		Scheduled: beforeScheduled,
	}
	before, err := json.Marshal(beforeDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling before diff output: %v", err))
	}
	diff["before"] = string(before)

	afterScheduled := false
	if after, exists := h.resources[afterDiffField]; exists {
		afterScheduled, ok = after.(bool)
		if !ok {
			panic(fmt.Sprintf("invalid afterScheduled value: cannot parse '%s' to bool",
				h.resources[afterDiffField]))
		}
	}

	afterDiffOutput := hostRebootDiffOutput{
		Scheduled: afterScheduled,
	}
	after, err := json.Marshal(afterDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling after diff output: %v", err))
	}
	diff["after"] = string(after)

	return diff
}

// isRebootScheduled checks if there is a scheduled reboot by querying systemd via D-Bus
func (h *HostReboot) isRebootScheduled(ctx context.Context) (bool, error) {
	// Connect to systemd D-Bus
	conn, err := h.dbusConstructor(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to connect to systemd D-Bus: %w", err)
	}
	defer conn.Close()

	// Check if there's a scheduled shutdown job
	// We look for shutdown.target or reboot.target in scheduled jobs
	// queries org.freedesktop.systemd1.Manager.ListJobs
	jobs, err := conn.ListJobsContext(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list systemd jobs: %w", err)
	}

	// Look for reboot or shutdown related jobs
	for _, job := range jobs {
		jobName := job.Unit
		// jobs name as mentioned in systemd documentation
		// https://www.freedesktop.org/software/systemd/man/latest/systemd.special.html
		if jobName == "reboot.target" || jobName == "shutdown.target" ||
			jobName == "poweroff.target" || jobName == "halt.target" {
			return true, nil
		}
	}

	// Also check for any active shutdown timers
	// queries org.freedesktop.systemd1.Manager.ListUnits
	timers, err := conn.ListUnitsContext(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list systemd units: %w", err)
	}

	for _, unit := range timers {
		if unit.LoadState == "loaded" && unit.ActiveState == "active" &&
			(unit.Name == "shutdown.timer" || unit.Name == "reboot.timer") {
			return true, nil
		}
	}

	// Check for systemd-shutdown processes
	return h.hasActiveShutdownProcess(ctx), nil
}

// hasActiveShutdownProcess checks if there are any active shutdown processes
// This is a fallback method to detect scheduled shutdowns
func (h *HostReboot) hasActiveShutdownProcess(ctx context.Context) bool {
	// Check if shutdown command is running or if there's a scheduled shutdown
	_, err := h.executor.CombinedOutputContext(ctx, "pgrep", "-f", "shutdown")
	if err == nil {
		h.logger.Debug("Found active shutdown process")
		return true
	}

	// Check for systemd-shutdown
	_, err = h.executor.CombinedOutputContext(ctx, "pgrep", "-f", "systemd-shutdown")
	if err == nil {
		h.logger.Debug("Found systemd-shutdown process")
		return true
	}

	// Check if there's a /run/systemd/shutdown/scheduled file
	_, err = h.executor.CombinedOutputContext(ctx, "test", "-f", "/run/systemd/shutdown/scheduled")
	if err == nil {
		h.logger.Debug("Found systemd shutdown scheduled file")
		return true
	}

	return false
}
