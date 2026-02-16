package operator

import (
	"context"
	"fmt"

	"github.com/trento-project/agent/internal/core/systemd"
)

const PacemakerDisableOperatorName = "pacemakerdisable"

type ServiceDisableOption Option[ServiceDisable]

// ServiceDisable operator disables a systemd unit.
//
// # Execution Phases
//
// - PLAN:
//   The operator connects to systemd and determines if the service is enabled.
//   The operation is skipped if the service is already disabled.
//
// - COMMIT:
//   It disables the systemd unit.
//
// - VERIFY:
//   The operator checks if the service is disabled after the commit phase.
//
// - ROLLBACK:
//   If an error occurs during the COMMIT or VERIFY phase, the service is enabled back again.

type ServiceDisable struct {
	baseOperator
	systemdLoader    systemd.Loader
	systemdConnector systemd.Systemd
	service          string
}

func WithCustomServiceDisableSystemdLoader(systemdLoader systemd.Loader) ServiceDisableOption {
	return func(sd *ServiceDisable) {
		sd.systemdLoader = systemdLoader
	}
}

func WithServiceToDisable(service string) ServiceDisableOption {
	return func(sd *ServiceDisable) {
		sd.service = service
	}
}

func NewServiceDisable(
	name string,
	arguments Arguments,
	operationID string,
	options Options[ServiceDisable],
) *Executor {
	serviceDisable := &ServiceDisable{
		baseOperator: newBaseOperator(
			name, operationID, arguments, options.BaseOperatorOptions...,
		),
		systemdLoader: systemd.NewDefaultSystemdLoader(),
	}

	for _, opt := range options.OperatorOptions {
		opt(serviceDisable)
	}

	return &Executor{
		phaser:      serviceDisable,
		operationID: operationID,
		logger:      serviceDisable.logger,
	}
}

func (sd *ServiceDisable) plan(ctx context.Context) (bool, error) {
	systemdConnector, err := sd.systemdLoader.NewSystemd(ctx, systemd.WithCustomLogger(sd.logger))
	if err != nil {
		sd.logger.Error("unable to initialize systemd connector", "error", err)
		return false, fmt.Errorf("unable to initialize systemd connector: %w", err)
	}
	sd.systemdConnector = systemdConnector

	serviceEnabled, err := sd.systemdConnector.IsEnabled(ctx, sd.service)
	if err != nil {
		sd.logger.Error("failed to check if service is enabled", "service", sd.service, "error", err)
		return false, fmt.Errorf("failed to check if %s service is enabled: %w", sd.service, err)
	}

	sd.resources[beforeDiffField] = serviceEnabled

	if !serviceEnabled {
		sd.logger.Info("service is already disabled, skipping operation", "service", sd.service)
		sd.resources[afterDiffField] = serviceEnabled
		return true, nil
	}
	return false, nil
}

func (sd *ServiceDisable) commit(ctx context.Context) error {
	if err := sd.systemdConnector.Disable(ctx, sd.service); err != nil {
		sd.logger.Error("failed to disable service", "service", sd.service, "error", err)
		return fmt.Errorf("failed to disable service %s: %w", sd.service, err)
	}

	return nil
}

func (sd *ServiceDisable) verify(ctx context.Context) error {
	serviceEnabled, err := sd.systemdConnector.IsEnabled(ctx, sd.service)
	if err != nil {
		sd.logger.Error("failed to check if service is enabled", "service", sd.service, "error", err)
		return fmt.Errorf("failed to check if service %s is enabled: %w", sd.service, err)
	}

	if serviceEnabled {
		sd.logger.Info("service is not disabled, rolling back", "service", sd.service)
		return fmt.Errorf("service %s is not disabled", sd.service)
	}

	sd.resources[afterDiffField] = serviceEnabled

	return nil
}

func (sd *ServiceDisable) rollback(ctx context.Context) error {
	return sd.systemdConnector.Enable(ctx, sd.service)
}

func (sd *ServiceDisable) operationDiff(_ context.Context) map[string]any {
	return computeOperationDiff(sd.resources)
}

func (sd *ServiceDisable) after(_ context.Context) {
	sd.systemdConnector.Close()
}
