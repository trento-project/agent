package operator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/trento-project/agent/internal/core/systemd"
)

const (
	PacemakerEnableOperatorName = "pacemakerenable"
	pacemakerServiceName        = "pacemaker.service"
)

type serviceEnablementDiffOutput struct {
	Enabled bool `json:"enabled"`
}

type ServiceEnableOption Option[ServiceEnable]

// ServiceEnable operator enables a systemd unit.
//
// # Execution Phases
//
// - PLAN:
//   The operator connects to systemd and determines if the service is enabled.
//   The operation is skipped if the service is already enabled.
//
// - COMMIT:
//   It enables the systemd unit.
//
// - VERIFY:
//   The operator checks if the service is enabled after the commit phase.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the service is disabled back again.

type ServiceEnable struct {
	baseOperator
	systemdLoader    systemd.Loader
	systemdConnector systemd.Systemd
	service          string
}

func WithCustomServiceEnableSystemdLoader(systemdLoader systemd.Loader) ServiceEnableOption {
	return func(se *ServiceEnable) {
		se.systemdLoader = systemdLoader
	}
}

func WithServiceToEnable(service string) ServiceEnableOption {
	return func(se *ServiceEnable) {
		se.service = service
	}
}

func NewServiceEnable(
	name string,
	arguments Arguments,
	operationID string,
	options Options[ServiceEnable],
) *Executor {
	serviceEnable := &ServiceEnable{
		baseOperator: newBaseOperator(
			name, operationID, arguments, options.BaseOperatorOptions...,
		),
		systemdLoader: systemd.NewDefaultSystemdLoader(),
	}

	for _, opt := range options.OperatorOptions {
		opt(serviceEnable)
	}

	return &Executor{
		phaser:      serviceEnable,
		operationID: operationID,
		logger:      serviceEnable.logger,
	}
}

func (se *ServiceEnable) plan(ctx context.Context) (bool, error) {
	systemdConnector, err := se.systemdLoader.NewSystemd(ctx, systemd.WithCustomLogger(se.logger))
	if err != nil {
		se.logger.Error("unable to initialize systemd connector", "error", err)
		return false, fmt.Errorf("unable to initialize systemd connector: %w", err)
	}
	se.systemdConnector = systemdConnector

	serviceEnabled, err := se.systemdConnector.IsEnabled(ctx, se.service)
	if err != nil {
		se.logger.Error("failed to check if service is enabled", "service", se.service, "error", err)
		return false, fmt.Errorf("failed to check if %s service is enabled: %w", se.service, err)
	}

	se.resources[beforeDiffField] = serviceEnabled

	if serviceEnabled {
		se.logger.Info("service already enabled, skipping operation", "service", se.service)
		se.resources[afterDiffField] = serviceEnabled
		return true, nil
	}

	return false, nil
}

func (se *ServiceEnable) commit(ctx context.Context) error {
	if err := se.systemdConnector.Enable(ctx, se.service); err != nil {
		se.logger.Error("failed to enable service", "service", se.service, "error", err)
		return fmt.Errorf("failed to enable service %s: %w", se.service, err)
	}
	return nil
}

func (se *ServiceEnable) verify(ctx context.Context) error {
	serviceEnabled, err := se.systemdConnector.IsEnabled(ctx, se.service)
	if err != nil {
		se.logger.Error("failed to check if service is enabled", "service", se.service, "error", err)
		return fmt.Errorf("failed to check if service %s is enabled: %w", se.service, err)
	}

	if !serviceEnabled {
		se.logger.Info("service %s is not enabled, rolling back", "service", se.service)
		return fmt.Errorf("service %s is not enabled", se.service)
	}

	se.resources[afterDiffField] = serviceEnabled

	return nil
}

func (se *ServiceEnable) rollback(ctx context.Context) error {
	return se.systemdConnector.Disable(ctx, se.service)
}

func (se *ServiceEnable) operationDiff(_ context.Context) map[string]any {
	return computeOperationDiff(se.resources)
}

func (se *ServiceEnable) after(_ context.Context) {
	se.systemdConnector.Close()
}

func computeOperationDiff(resources map[string]any) map[string]any {
	diff := make(map[string]any)

	beforeEnabled, ok := resources[beforeDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid beforeEnabled value: cannot parse '%s' to bool",
			resources[beforeDiffField]))
	}

	afterEnabled, ok := resources[afterDiffField].(bool)
	if !ok {
		panic(fmt.Sprintf("invalid afterEnabled value: cannot parse '%s' to bool",
			resources[afterDiffField]))
	}

	beforeDiffOutput := serviceEnablementDiffOutput{
		Enabled: beforeEnabled,
	}
	before, err := json.Marshal(beforeDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling before diff output: %v", err))
	}
	diff[beforeDiffField] = string(before)

	afterDiffOutput := serviceEnablementDiffOutput{
		Enabled: afterEnabled,
	}
	after, err := json.Marshal(afterDiffOutput)
	if err != nil {
		panic(fmt.Sprintf("error marshalling after diff output: %v", err))
	}
	diff[afterDiffField] = string(after)

	return diff
}
